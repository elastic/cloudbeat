# Test Environments Deployment

**Motivation**
To provide an easy and deterministic way to set up the latest cloud environment, ensuring proper monitoring and usability


**Prerequisite**

This project utilizes AWS and Elastic Cloud accounts. To ensure proper deployment and usage, it is essential to obtain appropriate licenses in compliance with the licensing terms and conditions provided by the respective service providers.

Follow the [prerequisites](/README.md#prerequisites) chapter of our main README.

## Environment Variables & EC API Key

To generate an Elastic Cloud token, you have two options:

1. Follow the [Cloud API Keys](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html) documentation for step-by-step instructions on generating the token.

2. If you are already familiar with the token generation process, you can directly access the [Elastic Cloud Keys](https://cloud.elastic.co/deployment-features/keys) page to generate the token.

Choose the method that is most convenient for you to obtain the Elastic Cloud token required for deployment.


Ensure that the following AWS credentials are defined:

- `AWS_ACCESS_KEY_ID`: Your AWS access key ID.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key.


To successfully deploy the environment, ensure that the following variables are provided as deployment parameters or exported as environment variables:

```bash
export TF_VAR_ec_api_key={TOKEN} # <-- should be replaced by Elastic Cloud TOKEN
export TF_VAR_stack_version=8.7.2-SNAPSHOT
export TF_VAR_ess_region=gcp-us-west2
```

## Modules

This project leverages a set of foundational modules specifically designed for [cloud deployment](../cloud/modules/).

### EC2

**aws_ec2_for_kspm** - This module facilitates the deployment of an EC2 instance with a Kubernetes cluster using the kind tool. The deployment process relies on a customized image that includes the necessary components for running kind.

**aws_ec2_for_cspm** - This module facilitates the deployment of an EC2 instance for CSPM.

Please note that the customized image is currently available in the following regions: **eu-west-1** and **eu-west-3**. Therefore, ensure that you deploy this module in one of these regions to leverage the customized image effectively.

**Module variables (CSPM / KSPM)**

| Variable  | Default Value | Comment |
|:-------------:|:-------------:|:------------|
| region      |   eu-west-1   | AWS EC2 deployment region |



### Elastic Cloud

**ec_deployment** - This module facilitates the deployment of Elastic Cloud instance.

| Variable  | Default Value | Comment |
|:-------------:|:-------------:|:------------|
| ec_api_key    |   None   | The API key for Elastic Cloud can also be defined using the `TF_VAR_ec_api_key` environment variable |
| ess_region    | gcp-us-west2 | The ESS deployment region can also be defined using the `TF_VAR_stack_version` environment variable|
| stack_version | latest | The ELK stack version can also be defined using the `TF_VAR_stack_version` environment variable |
| pin_version   | None | Optional: The ELK pin version (docker tag override) can also be defined using the `TF_VAR_pin_version` environment variable |

## Execution

To execute the full project, encompassing the deployment of an EC2 instance, setting up a Kubernetes cluster using kind, and deploying Elastic Cloud, follow the steps outlined below

- Initiate the project

```bash
cd test-environments
terraform init
```

- Deploy test environment

```bash
terraform apply --auto-approve -var="deployment_name=dev-env"
```

For development purposes, it is possible to deploy each module separately, allowing for focused and independent development and testing. Each module within the project represents a specific component or functionality and can be deployed individually to streamline the development process.

Below are examples demonstrating how to execute individual modules separately:

- EC2 for CSPM

```bash
terraform apply --auto-approve -target "module.aws_ec2_for_cspm"
```

- EC2 + Kind Kubernetes (KSPM)

```bash
terraform apply --auto-approve -target "module.aws_ec2_for_kspm"
```

- EC Deployment

```bash
terraform apply --auto-approve -target "module.ec_deployment"
```

BC version

```bash
terraform apply --auto-approve -var="stack_version=8.12.0" -var="pin_version=8.12.0-9f05a310" -target "module.ec_deployment"
```


- EKS Deployment

```bash
terraform apply --auto-approve -target "module.eks"
```

## Environment Cleanup

To destroy local environment use

``` bash
terraform destroy -var="region=eu-west-1"
```


To destroy the environment provisioned using the Sanity job, follow these steps:

1. [Download](https://s3.console.aws.amazon.com/s3/buckets/tf-state-bucket-test-infra?region=eu-west-3&tab=objects) the Terraform state file to the [test-environments](../test-environments/) folder.
2. Rename the state file, for example, `terraform-sanity.tfstate`.
3. Run the following command:

``` bash
terraform destroy -var="region=eu-west-1" -state terraform-sanity.tfstate
```