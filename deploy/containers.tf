# local variables for use in container creation
locals {
  version                = formatdate("YYYYMMDDhhmmss", timestamp())
  upload_processor_image = "gcr.io/${var.gcp_project_id}/upload-processor:${local.version}"
  web_interface_image    = "gcr.io/${var.gcp_project_id}/web-interface:${local.version}"
}

# google container registry for container images
resource "google_container_registry" "this" {
}

# cli commands to build the docker containers for the upload processing and push to the google container registry
resource "null_resource" "upload_processor" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    working_dir = path.module
    command     = "gcloud auth configure-docker && docker build -t ${local.upload_processor_image} ../upload-processor/ && docker push ${local.upload_processor_image}"
  }
}

# cli commands to build the docker containers for the web interface and push to the google container registry
resource "null_resource" "web_interface" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    working_dir = path.module
    command     = "gcloud auth configure-docker && docker build -t ${local.web_interface_image} ../web-interface/ && docker push ${local.web_interface_image}"
  }
}

