# Deploy devnet on aws using terraform

## Prerequisites
- Install polygon-edge `make build-edge`
- Build the server for linux `make build-server GOOS=linux` (the server binary will later be copied to the remote server and run there so the GOOS and GOARCH has to match)
- Build the accounts tool for linux (since it's going to be run on the remote machine) `make tools-account GOOS=linux`
- Build the staking contract `make build-staking-contract`
- Install aws cli tool and run `aws configure`, copy your Access Key ID and Secret Access Key from the aws console.
- Install session manager plugin for AWS CLI
- Install terraform and run `terraform login`.
- `jq` tool has to be present on your machine

## Provision the AWS resources using terraform and deploy the nodes

Terraform requires one variable

Run commands:
- `cd ./deployment/devnet`
- `terraform init & terraform apply` 
- `./deploy.sh`

You can configure the deployment options using terraform variables like so: `terraform apply -var <key>=<value>` or `terraform apply -var-file="<filename>.tfvars"`
Check out [variables.tf](./devnet/variables.tf) to see what variables you can provide in order to customize the deployment.

