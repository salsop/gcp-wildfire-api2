# region to deploy the regional resources to
variable "gcp_region" {
  description = "region for deployment of all regional resources"
  default     = "europe-west2"
}

# gcp project id for all deployment of resources
variable "gcp_project_id" {
  description = "google project id for the deployment of all resources"
}

# wildfire api global or regional cloud to use.
variable "wildfire_api_portal" {
  description = "wildfire api portal"
  default     = "wildfire.paloaltonetworks.com"
}

# wildfire api key for authentication to the wildfire api
variable "wildfire_api_key" {
  description = "wildfire api key for authentication"
}