package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	storage "cloud.google.com/go/storage"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var WildfireAPIKey string
var err error

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

/// google storage object information provided in the Data of the PubSubMessage
type GCSEvent struct {
	Kind                    string                 `json:"kind"`
	ID                      string                 `json:"id"`
	SelfLink                string                 `json:"selfLink"`
	Name                    string                 `json:"name"`
	Bucket                  string                 `json:"bucket"`
	Generation              string                 `json:"generation"`
	Metageneration          string                 `json:"metageneration"`
	ContentType             string                 `json:"contentType"`
	TimeCreated             time.Time              `json:"timeCreated"`
	Updated                 time.Time              `json:"updated"`
	TemporaryHold           bool                   `json:"temporaryHold"`
	EventBasedHold          bool                   `json:"eventBasedHold"`
	RetentionExpirationTime time.Time              `json:"retentionExpirationTime"`
	StorageClass            string                 `json:"storageClass"`
	TimeStorageClassUpdated time.Time              `json:"timeStorageClassUpdated"`
	Size                    string                 `json:"size"`
	MD5Hash                 string                 `json:"md5Hash"`
	MediaLink               string                 `json:"mediaLink"`
	ContentEncoding         string                 `json:"contentEncoding"`
	ContentDisposition      string                 `json:"contentDisposition"`
	CacheControl            string                 `json:"cacheControl"`
	Metadata                map[string]interface{} `json:"metadata"`
	CRC32C                  string                 `json:"crc32c"`
	ComponentCount          int                    `json:"componentCount"`
	Etag                    string                 `json:"etag"`
	CustomerEncryption      struct {
		EncryptionAlgorithm string `json:"encryptionAlgorithm"`
		KeySha256           string `json:"keySha256"`
	}
	KMSKeyName    string `json:"kmsKeyName"`
	ResourceState string `json:"resourceState"`
}

// Wildfire verdict response object
type wildfireVerdict struct {
	Sha256  string `xml:"get-verdict-info>sha256"`
	Verdict string `xml:"get-verdict-info>verdict"`
	Md5     string `xml:"get-verdict-info>md5"`
	Error   string `xml:"error-message"`
}

// Wildfire upload response object
type wildfireUpload struct {
	Error string `xml:"error-message"`
}

// Gets a secret from the Google Secrets Manager and returns it as a string
//   name format projects/project-id/secrets/secret-name/versions/latest
func getSecretValue(name string) (string,error){
	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create secret manager client - %v \n", err)
		return "",err
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Printf("failed to access secret version - %v - %v \n", name, err)
		return "",err
	}

	return string(result.Payload.Data), nil
}

// Check MD5 Hash in Wildfire Database
//   Returns error and verdict result
func checkWildfireVerdictByMD5(md5Hash string) string {
	// get wildfire api portal and key from GCP secrets manager
	wildfireAPIPortal := os.Getenv("WILDFIRE_API_PORTAL")

	// make api call to wildfire to get verdict
	data := url.Values{}
	data.Set("apikey", WildfireAPIKey)
	data.Set("hash", md5Hash)

	fullURL := "https://" + wildfireAPIPortal + "/publicapi/get/verdict"

	resp, err := http.PostForm(fullURL, data)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var verdict wildfireVerdict

	ByteValue, _ := ioutil.ReadAll(resp.Body)

	err = xml.Unmarshal(ByteValue, &verdict)
	if err != nil {
		fmt.Println(err)
	}

	// check for all possible verdict responses
	switch verdict.Verdict {
	case "0":
		return "benign"
	case "1":
		return "malware"
	case "2":
		return "grayware"
	case "4":
		return "phishing"
	case "5":
		return "c2"
	case "-100":
		return "pending, the sample exists, but there is currently no verdict (applicable to file analysis only)"
	case "-101":
		return "unknown - error -101"
	case "-102":
		return "unknown - cannot find sample record in the database"
	case "-103":
		return "unknown - invalid hash value"
	default:
		return "unknown - no verdict"
	}

	return "" // Return No Error
}

// decode the GCS md5 value to a MD5 hash string
func decodeGCSMD5Value(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	x := hex.EncodeToString(data)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return x
}

// upload file contents to wildfire for analysis
func uploadFileToWildfire(filename, contents string) error {
	// get wildfire api portal and key from GCP secrets manager
	wildfireAPIPortal := os.Getenv("WILDFIRE_API_PORTAL")

	url := "https://" + wildfireAPIPortal + "/publicapi/submit/file"

	body := &bytes.Buffer{}

	file := strings.NewReader(contents)

	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(filename))
	_, err := io.Copy(part, file)
	if err != nil {
		log.Print(err)
	}

	err = writer.WriteField("apikey", WildfireAPIKey)
	if err != nil {
		log.Print(err)
	}

	writer.Close()

	r, _ := http.NewRequest("POST", url, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Print(err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var uploadResponse wildfireUpload

	err = xml.Unmarshal(content, &uploadResponse)
	if err != nil {
		fmt.Println(err)
	}

	if uploadResponse.Error != "" {
		fmt.Printf("error: %v \n", uploadResponse.Error)
		return fmt.Errorf("%v", uploadResponse.Error)
	}

	return nil
}

func moveFile(srcBucket, dstBucket, objName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	src := client.Bucket(srcBucket).Object(objName)
	dst := client.Bucket(dstBucket).Object(objName)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", objName, objName, err)
	}
	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", objName, err)
	}
	fmt.Printf("Blob %v moved to %v.\n", objName, dstBucket)
	return nil
}

// gets the contents of the gcs bucket object and returns it a
func getFileContents(bucket, object string) string {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	body, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

// CloudFunction Entrypoint
func GCSFileUploaded(ctx context.Context, e GCSEvent) error {
	// return nothing
	return nil
}

func PubSubProcessor(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Printf("Message Received:\n %v", string(m.Message.Data))
	var File GCSEvent
	err = json.Unmarshal(m.Message.Data, &File)
	if err != nil {
		log.Printf("error unmarshalling data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	md5Hash := decodeGCSMD5Value(File.MD5Hash)
	log.Printf("object: %v \n", File.Name)
	log.Printf("bucket: %v \n", File.Bucket)
	log.Printf("md5 hash: %v \n", md5Hash)

	quarantineBucket := os.Getenv("QUARANTINE_BUCKET")
	cleanBucket := os.Getenv("SCANNED_BUCKET")
	unsupportedBucket := os.Getenv("UNSUPPORTED_BUCKET")

	// Check if Wildfire has a verdict for the file by MD5 Hash
	verdict := checkWildfireVerdictByMD5(md5Hash)

	fmt.Printf("md5 hash: %v, verdict: %v \n", md5Hash, verdict)

	switch verdict {

	// if standard verdict, then update file metadata with result
	case "benign":
		err := moveFile(File.Bucket, cleanBucket, File.Name)
		if err != nil {
			log.Print(err)
		}

	case "malware", "phishing":
		err := moveFile(File.Bucket, quarantineBucket, File.Name)
		if err != nil {
			log.Print(err)
		}

		// if not a standard verdict, upload the file for analysis
	default:
		contents := getFileContents(File.Bucket, File.Name)
		fmt.Printf("uploading %v (%v) to wildfire for analysis \n", File.Name, len(contents))
		err := uploadFileToWildfire(File.Name, contents)
		if err != nil {
			if strings.Contains(fmt.Sprint(err), "Unsupport File type with sample") {
				moveFile(File.Bucket, unsupportedBucket, File.Name)
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for {
			time.Sleep(20 * time.Second)
			verdict := checkWildfireVerdictByMD5(md5Hash)

			switch verdict {
			case "malware", "phishing":
				fmt.Printf("md5 hash: %v, verdict: %v \n", md5Hash, verdict)
				err := moveFile(File.Bucket, quarantineBucket, File.Name)
				if err != nil {
					log.Print(err)
				}
				w.WriteHeader(http.StatusOK)
				return

			case "benign":
				fmt.Printf("md5 hash: %v, verdict: %v \n", md5Hash, verdict)
				err := moveFile(File.Bucket, cleanBucket, File.Name)
				if err != nil {
					log.Printf("error moving file: %v \n", err)
				}
				w.WriteHeader(http.StatusOK)
				return

			default:
				fmt.Printf("waiting for analysis...(current verdict: %v) \n", verdict)
			}
		}
	}
}

func main() {
	// get wildfire api key from secret manager
	projectId := os.Getenv("GCP_PROJECT")
	WildfireAPIKey, err = getSecretValue("projects/" + projectId + "/secrets/wildfire_api_key/versions/latest")
	if err != nil {
		log.Printf("cannot read secret value - %v \n", err)
	} else
	{
	    log.Printf("loaded wildfire_api_key secret")
	}

	http.HandleFunc("/", PubSubProcessor)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
