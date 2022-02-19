# random prefix for naming of resources
resource "random_pet" "unique" {
}

# output of web interface url
output "web_interface" {
  value = <<EOF



  =================================================================================
  >>>>>>>>>>>>>>>>>>>>>>>>>>>>>  Deployment Completed <<<<<<<<<<<<<<<<<<<<<<<<<<<<<
  =================================================================================

  To access the website to upload files open a browser and visit the following URL:

    ${google_cloud_run_service.web_interface.status[0].url}

  Once uploaded you can check the following Google Cloud Storage buckets:

    ${google_storage_bucket.upload.name}      - Uploaded files.
    ${google_storage_bucket.unsupported.name} - Files that are not supported by WildFire.
    ${google_storage_bucket.quarantine.name}  - Files found to be "malicious".
    ${google_storage_bucket.scanned.name}     - Files found to be "benign".

  To get detailed logs you can review the Cloud Run deployments:

    ${google_cloud_run_service.web_interface.name} - Web interface and creating of file in the bucket.
    ${google_cloud_run_service.upload_processor.name} - Processing of PubSub messages and uploaded files.


EOF
}

