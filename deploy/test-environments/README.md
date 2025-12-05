# Test Environments Deployment

**Motivation**
To provide an easy and deterministic way to set up the latest cloud environment, ensuring proper monitoring and usability


**Prerequisite**

This project utilizes AWS, Elastic Cloud, Azure, and GCP accounts. To ensure proper deployment and usage, you need to obtain appropriate licenses and credentials in compliance with the licensing terms and conditions provided by the respective service providers.

Follow the [prerequisites](/README.md#prerequisites) chapter of our main README.

## Environment Variables & EC API Key

To generate an Elastic Cloud token, you have two options:

1. Follow the [Cloud API Keys](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html) documentation for step-by-step instructions on generating the token.

2. If you are already familiar with the token generation process, you can directly access the [Elastic Cloud Keys](https://cloud.elastic.co/deployment-features/keys) page to generate the token.

Choose the method that is most convenient for you to obtain the Elastic Cloud token required for deployment.


For AWS:

Ensure that the following AWS credentials are defined:

- `AWS_ACCESS_KEY_ID`: Your AWS access key ID.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key.

For GCP:

Ensure that you have your GCP service account key file. This file is usually stored at a path like ~/.config/gcloud, but the exact location may vary.

For Azure:

Ensure that you are logged in to Azure using:

```bash
az login
```

To successfully deploy the environment, ensure that the following variables are provided as deployment parameters or exported as environment variables:

```bash
export TF_VAR_ec_api_key={TOKEN} # <-- should be replaced by Elastic Cloud TOKEN
export TF_VAR_stack_version=8.16.0-SNAPSHOT
export TF_VAR_ess_region=production-cft  # or staging-aws, qa-azure, etc.
```

## Directory Structure

### elk-stack

This directory handles the deployment of the Elastic Stack. It includes:

- Deployment Types: `deployment` and `project`, defined by the var.serverless_mode key.
- Required Variable: `ec_api_key`. It is also recommended to provide `deployment_name` as an input parameter during development.

**ec_deployment** - This module facilitates the deployment of Elastic Cloud instance.

| Variable  | Default Value | Comment |
|:-------------:|:-------------:|:------------|
| ec_api_key    |   None   | The API key for Elastic Cloud can also be defined using the `TF_VAR_ec_api_key` environment variable |
| ess_region    | gcp-us-west2 | The ESS deployment region can also be defined using the `TF_VAR_stack_version` environment variable|
| stack_version | latest | The ELK stack version can also be defined using the `TF_VAR_stack_version` environment variable |
| pin_version   | None | Optional: The ELK pin version (docker tag override) can also be defined using the `TF_VAR_pin_version` environment variable |


### cis

This directory is responsible for provisioning EC2 machines and EKS clusters related to CSPM and KSPM.

**aws_ec2_for_kspm** - This module facilitates the deployment of an EC2 instance with a Kubernetes cluster using the kind tool. The deployment process relies on a customized image that includes the necessary components for running kind.

**aws_ec2_for_cspm** - This module facilitates the deployment of an EC2 instance for CSPM.

Please note that the customized image is currently available in the following regions: **eu-west-1** and **eu-west-3**. Therefore, ensure that you deploy this module in one of these regions to leverage the customized image effectively.

### cdr

This directory includes modules for provisioning infrastructure for CDR, including:

- GCP VM (requires gcp_project_id as an input variable)
- Azure VM
- AWS EC2 for CloudTrail
- Additional EC2 for asset inventory

### Modules

All projects utilize a set of foundational modules specifically designed for [cloud deployment](./modules/).


## Execution

There is no single Terraform command to execute the full project. Instead, each module can be executed separately using Terraform commands. The scripts provided in the project are responsible for managing the execution of the entire setup.

### Full Project Execution

The full project execution is managed by scripts, not by Terraform directly. Use the following scripts to handle the deployment process:

- `manage_infrastructure.sh`: This script manages Terraform provisioning with commands for {elk-stack|cis|cdr|all} {apply|destroy|output|upload-state}.
The following command applies all Terraform configurations for the elk-stack, cis, and cdr directories:
```bash
./manage_infrastructure.sh all apply
```
The following command destroys all Terraform configurations for the elk-stack, cis, and cdr directories:

```bash
./manage_infrastructure.sh all destroy
```

The following command retrieves outputs from all deployed environments:
```bash
./manage_infrastructure.sh all output
```


### Running Individual Modules

- Elastic Stack

```bash
cd elk-stack
terraform init
terraform apply --auto-approve
```

- CIS modules

```bash
cd cis
terraform init
terraform apply --auto-approve
```

- Specific module in CIS

```bash
cd cis
terraform apply --auto-approve -target "module.aws_ec2_for_kspm"
```

- CDR modules

```bash
cd cdr
terraform init
terraform apply --auto-approve
```

## Environment Cleanup

To destroy local environment use

```bash
./manage_infrastructure.sh all destroy
```
