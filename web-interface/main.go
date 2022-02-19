package main

import (
	"context"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
)

type responsePage struct {
	ResponseTitle string
	ResponseText  string
	ResponseStyle string
}

// creates a gcs object (file) with then name specified from the passed slice of bytes
func createGoogleStorageFile(name string, file []byte) error {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Print("no bucket specified")
		return errors.New("no bucket specified")
	}

	log.Printf("creating %v in %v", name, bucketName)

	ctx := context.Background()
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Print(err)
		return err
	}
	gcsObject := gcsClient.Bucket(bucketName).Object(name).NewWriter(ctx)
	defer gcsObject.Close()

	_, err = gcsObject.Write(file)
	if err != nil {
		log.Print(err)
		return err
	}

	log.Printf("%v created in %v", name, bucketName)

	return nil
}

func main() {
	// check if PORT is specified
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// show admin what port we're listening on
	log.Printf("starting web server on :" + port)

	// check if GCS_BUCKET_NAME is specified
	if os.Getenv("GCS_BUCKET_NAME") == "" {
		log.Fatal("error no bucket configured, configure environment variable GCS_BUCKET_NAME")
	}

	// handlefunc for /
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// log the method and URL
		log.Printf("Method: %v for %v", r.Method, r.URL)

		// GET method logic
		if r.Method == "GET" {
			tmplUploadPage, err := template.ParseFiles("upload-page.html")
			if err != nil {
				log.Fatal(err)
			}

			tmplUploadPage.Execute(w, nil)
		}

		// POST method logic
		if r.Method == "POST" {
			tmplResponsePage, err := template.ParseFiles("response-page.html")
			if err != nil {
				log.Fatal(err)
			}
			file, handler, err := r.FormFile("file")
			if err != nil {
				log.Print(err)
				tmplResponsePage.Execute(w, responsePage{ResponseTitle: "File upload failed", ResponseText: "I'm sorry there was a problem with the file upload, please ensure you selected a file to upload.", ResponseStyle: "alert-danger"})

			} else {

				log.Printf("File Name: %v", handler.Filename)
				log.Printf("File Size: %v", handler.Size)

				fileContents, err := io.ReadAll(file)
				if err != nil {
					log.Print(err)
				}

				err = createGoogleStorageFile(handler.Filename, fileContents)
				if err != nil {
					log.Print(err)
					tmplResponsePage.Execute(w, responsePage{ResponseTitle: "File upload failed", ResponseText: "I'm sorry there was a problem with the file upload, please try again.", ResponseStyle: "alert-danger"})
				} else {
					tmplResponsePage.Execute(w, responsePage{ResponseTitle: "File upload completed", ResponseText: "Thank you, your file has been received and is being processed.", ResponseStyle: "alert-success"})
				}
			}
		}
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
