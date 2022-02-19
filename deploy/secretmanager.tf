# create secret for wildfire api key
resource "google_secret_manager_secret" "wildfire_api_key" {
  project   = var.gcp_project_id
  secret_id = "wildfire_api_key"
  replication {
    automatic = true
  }
}

# set the value for the secret
resource "google_secret_manager_secret_version" "wildfire_api_key" {
  secret      = google_secret_manager_secret.wildfire_api_key.id
  secret_data = var.wildfire_api_key
}

# give the cloud function access to the wildfire api key secret
resource "google_secret_manager_secret_iam_binding" "wildfire_api_key" {
  members   = ["serviceAccount:${google_cloud_run_service.upload_processor.template[0].spec[0].service_account_name}"]
  role      = "roles/secretmanager.secretAccessor"
  secret_id = google_secret_manager_secret.wildfire_api_key.id
}
