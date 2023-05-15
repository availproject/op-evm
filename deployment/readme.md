# Deploy devnet on aws using terraform

## Prerequisites
- Install aws cli tool and run `aws configure`, copy your Access Key ID and Secret Access Key from the aws console.
- Install session manager plugin for AWS CLI
- Install terraform and run `terraform login`.

## Provision the AWS resources using terraform and deploy the nodes

Terraform requires `github_token` variable to get the release from the private repo.

Run commands:
- `cd ./deployment/devnet`
- `terraform init`
- `terraform apply -var github_token=$G_TOKEN` 

You can configure the deployment options using terraform variables like so: `terraform apply -var <key>=<value>` or `terraform apply -var-file="<filename>.tfvars"`
Check out [variables.tf](./devnet/variables.tf) to see what variables you can provide in order to customize the deployment.

### Debugging instances

Install session manager plugin for AWS CLI (On macOS)
___
- `curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip" -o "sessionmanager-bundle.zip"`
- `unzip sessionmanager-bundle.zip`
- `sudo ./sessionmanager-bundle/install -i /usr/local/sessionmanagerplugin -b /usr/local/bin/session-manager-plugin`

Get the private key from terraform
___
- `terraform output --raw pk > key.pem`
- `chmod 400 key.pem`

Configure aws proxy options and connect using ssh
___
- `vi ~/.ssh/config`
- Add the following lines: 
  ```
  host i-* mi-*
  ProxyCommand sh -c "aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'"
  ```
- `chmod 600 ~/.ssh/config`
- `ssh -i key.pem ec2-user@[INSTANCE-ID]`

For more info check:
- https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html
- https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-troubleshooting.html#plugin-not-found
- https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html
