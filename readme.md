1. Create Project and Assign Billing Account
2. Enable the required service APIs for the project:

```shell
gcloud services enable \
  container.googleapis.com \
  secretmanager.googleapis.com \
  run.googleapis.com
```

3. Clone the Repo:

```shell
git clone https://github.com/salop/gcp-wildfire-api2
```

4. Create a `terraform.tfvars` file with the required settings

```shell
cd deploy
```
```
export gcp_project=xxxx
```
```
export wildfire_api_key=yyyy
```
```
cat <<EOF > terraform.tfvars
gcp_project_id   = "$gcp_project"
wildfire_api_key = "$wildfire_api_key"
EOF
```

4. Use Terraform to deploy to GCP

```shell
terraform init
```

```shell
terraform apply
```
