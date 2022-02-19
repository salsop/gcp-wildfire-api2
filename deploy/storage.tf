# google storage bucket for uploads of files
resource "google_storage_bucket" "upload" {
  name          = "${random_pet.unique.id}-upload"
  location      = var.gcp_region
  force_destroy = true
}

# google storage bucket for files found to be malicious or grayware
resource "google_storage_bucket" "quarantine" {
  name          = "${random_pet.unique.id}-quarentine"
  location      = var.gcp_region
  force_destroy = true
}

# google storage bucket for files found to be benign
resource "google_storage_bucket" "scanned" {
  name          = "${random_pet.unique.id}-scanned"
  location      = var.gcp_region
  force_destroy = true
}

# google storage bucket for files of unsupported file types
resource "google_storage_bucket" "unsupported" {
  name          = "${random_pet.unique.id}-unsupported"
  location      = var.gcp_region
  force_destroy = true
}
