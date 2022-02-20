# google pubsub topic for messages to be posted when new objects are created in the upload bucket
resource "google_pubsub_topic" "upload" {
  name = "${random_pet.unique.id}-upload"
}

# service account for allowing the pubsub subcrription to call the cloud run upload processing web service
resource "google_service_account" "upload_subscription" {
  account_id = "upload-subscription"
}

# assign the iam role for the service account to call the cloud run upload processing web service.
resource "google_cloud_run_service_iam_member" "member" {
  location = var.gcp_region
  service  = google_cloud_run_service.upload_processor.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.upload_subscription.email}"
}

# google pubsub subscription for push messages to be delivered to cloud run for processing
resource "google_pubsub_subscription" "upload" {
  name                 = "${random_pet.unique.id}-upload"
  topic                = google_pubsub_topic.upload.id
  ack_deadline_seconds = 600

  expiration_policy {
    ttl = "" # never expires
  }

  retry_policy {
    # default values
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  push_config {
    push_endpoint = google_cloud_run_service.upload_processor.status[0].url
    oidc_token {
      service_account_email = google_service_account.upload_subscription.email
    }
  }
}

# get project information
data "google_project" "this" {
}

# give the storage account the role to allow it to publish to the pubsub topic
resource "google_pubsub_topic_iam_binding" "upload" {
  topic   = google_pubsub_topic.upload.id
  role    = "roles/pubsub.publisher"
  members = ["serviceAccount:service-${data.google_project.this.number}@gs-project-accounts.iam.gserviceaccount.com"]
}

# google pubsub to cloud run push notification
resource "google_storage_notification" "notification" {
  bucket         = google_storage_bucket.upload.name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.upload.id
  event_types    = ["OBJECT_FINALIZE"]
  depends_on = [
    google_pubsub_topic_iam_binding.upload
  ]
}






