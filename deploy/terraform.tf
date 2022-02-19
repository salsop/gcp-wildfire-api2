
# fixed versions of the providers to avoid breaking changes
terraform {
  required_providers {
    google = {
      version = "~> 4.11"
    }
    random = {
      version = "~> 3.1"
    }
    null = {
      version = "~> 3.1"
    }
  }
}

# set the google provide to use the same project for all resources
provider "google" {
  project = var.gcp_project_id
}
