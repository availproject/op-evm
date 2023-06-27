# Deploy Development Network (DevNet) on AWS using Terraform

## Introduction
This documentation provides instructions on how to deploy a Development Network (DevNet) on AWS using Terraform.
The `devnet` directory consists of reusable [Terraform modules](https://www.terraform.io/language/modules).
The `nets` directory holds specific configurations for individual networks.
Said another way, the `nets` are instances of `devnet`.

## Prerequisites
Before proceeding with the deployment, ensure the following prerequisites are met:
- Install the AWS CLI tool and run `aws configure` to set up your Access Key ID and Secret Access Key obtained from the AWS console.
- Install the Session Manager plugin for AWS CLI.
- Install Terraform.

## Provisioning the Avail Network
To provision the Avail Network, follow these steps:
- Deploy your avail network.
- If your Avail Network is publicly available, pass the `avail_hostname` and `avail_port` variables to the Terraform script bellow using the `-var` or `-var-file` arguments.
- If your Avail Network is private and located in the same region and AWS account as this deployment, use the `avail_peer` variable to configure the peering. Typically, the `route53_zone_private_id`, `route_table_private_ids` and `vpc_id` will be outputted from the Avail deployment Terraform script.

## Provisioning AWS Resources and Deploying the Nodes
To provision the necessary AWS resources and deploy the nodes, perform the following steps:

1. The Terraform script requires the `github_token` variable to access the release from the private repository.
2. Open a terminal and navigate to the `./deployment/devnet` directory.
3. Initialize Terraform by running `terraform init`.
4. Apply the Terraform configuration using the command `terraform apply -var github_token=$G_TOKEN`.
  - You can customize the deployment options by specifying Terraform variables in the format `terraform apply -var <key>=<value>` or `terraform apply -var-file="<tfvars-filename>"`.
  - For available variables to customize the deployment, refer to the `devnet/variables.tf` file.

## Deploying Your Own Network
To deploy your own DevNet for testing purposes, follow these steps:
1. If you prefer not to push your configuration to the git repository, you have the option to create a Terraform module within the `nets/private` folder. This folder is specifically ignored by git.
2. If you wish to persist your configuration to git, create a Terraform module under the `nets/<devnet-name>` folder.
3. Inside the `nets/private` folder, create a file named `main.tf`.
4. Use the provided template below to fill in the required details:

```terraform
terraform {
  backend "s3" {
    bucket = "availsl-tf-states"
    key    = "state/op-evm/<deployment-name>"
    region = "<region>"
  }
}

module "devnet" {
  source          = "../../devnet"
  deployment_name = "<deployment-name>"
  region          = "<region>"
  base_ami        = "<ami>" # Latest ubuntu ami 
  avail_hostname  = "internal-rpc.testnetsl.avail.private"
  release         = "v0.1.0" # Use latest release
  avail_peer      = {
    route53_zone_private_id = "<route53-zone>"
    route_table_private_ids = [ 
      "<route-table-1>",
      "<route-table-2>",
      "<route-table-3>",
    ]
    vpc_id = "<vpc-id>"
  }
  github_token = "<github-token>"
}
```

## Debugging Instances
To debug instances in the DevNet, follow these steps:

### Installing the Session Manager Plugin for AWS CLI (macOS)

Run the following commands in the terminal:
```shell
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip" -o "sessionmanager-bundle.zip"
unzip sessionmanager-bundle.zip
sudo ./sessionmanager-bundle/install -i /usr/local/sessionmanagerplugin -b /usr/local/bin/session-manager-plugin
```

### Getting the Private Key from Terraform
1. Retrieve the private key from Terraform by running the command `terraform output --raw ssh_pk > key.pem`.
2. Set the correct permissions for the private key file using `chmod 600 key.pem`.

### Configuring AWS Proxy Options and Connecting via SSH
1. Open the `~/.ssh/config` file using a text editor.
2. Add the following lines to the file:
  ```
  host i-* mi-*
  ProxyCommand sh -c "aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'"
  ```
3. Save the changes and set the correct permissions for the `~/.ssh/config` file using `chmod 600 ~/.ssh/config`.
4. Connect to the instance using SSH with the following command (replace `[INSTANCE-ID]` with the actual instance ID):

### For more information, refer to the following resources
- [AWS Session Manager Plugin Installation Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)
- [AWS Session Manager Troubleshooting Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-troubleshooting.html#plugin-not-found)
- [AWS Session Manager Getting Started Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html)

## Forwarding Avail Explorer to Localhost using SSH Proxy
1. Open a new console and run the command `ssh -N -L 8888:internal-rpc.testnet04.avail.private:80 -i key.pem ubuntu@<instance-id>`.
2. In another console, run the command `ssh -N -L 9944:internal-rpc.testnet04.avail.private:8546 -i key.pem ubuntu@<instance-id>`.
3. Open your browser and access `localhost:8888`.
4. In the Avail Explorer, select the **Local Node** option (with address: `127.0.0.1:9944`) in the networks settings menu, located under development, and press **Switch**.
5. Explore the blocks in the Avail Explorer!
