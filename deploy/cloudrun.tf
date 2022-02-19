# google cloud run service for processing uploads
resource "google_cloud_run_service" "upload_processor" {
  location = var.gcp_region
  name     = "${random_pet.unique.id}-upload-processor"

  template {
    spec {
      timeout_seconds = 300 # the timeout before the message is replied back with a 504 is returned

      containers {
        image = local.upload_processor_image

        env {
          name  = "WILDFIRE_API_PORTAL"
          value = var.wildfire_api_portal
        }
        env {
          name  = "GCP_PROJECT"
          value = var.gcp_project_id
        }
        env {
          name  = "QUARANTINE_BUCKET"
          value = google_storage_bucket.quarantine.name
        }
        env {
          name  = "SCANNED_BUCKET"
          value = google_storage_bucket.scanned.name
        }
        env {
          name  = "UNSUPPORTED_BUCKET"
          value = google_storage_bucket.unsupported.name
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    null_resource.upload_processor,
  ]
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

# google cloud run service for web interface for uploads
resource "google_cloud_run_service" "web_interface" {
  location = var.gcp_region
  name     = "${random_pet.unique.id}-web-interface"

  template {
    spec {
      containers {
        image = local.web_interface_image

        env {
          name  = "GCS_BUCKET_NAME"
          value = google_storage_bucket.upload.name
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    null_resource.web_interface,
  ]

}

# assign policy to allow anyone to access the cloud run web interface
resource "google_cloud_run_service_iam_policy" "noauth" {
  location    = var.gcp_region
  service     = google_cloud_run_service.web_interface.name
  policy_data = data.google_iam_policy.noauth.policy_data
}