# Test Environments Deployment

**Motivation**
To provide an easy and deterministic way to set up the latest cloud environment, ensuring proper monitoring and usability


**Prerequisite**

This project utilizes AWS and Elastic Cloud accounts. To ensure proper deployment and usage, it is essential to obtain appropriate licenses in compliance with the licensing terms and conditions provided by the respective service providers. 

Follow the [prerequisites](/README.md#prerequisites) chapter of our main README.

## Environment Variables

To successfully deploy the environment, ensure that the following variables are provided as deployment parameters or exported as environment variables:

```bash
export TF_VAR_ec_api_key={TOKEN} # <-- should be replaced by Elastic Cloud TOKEN
export TF_VAR_stack_version=8.6.0-SNAPSHOT
export TF_VAR_ess_region=gcp-us-west2
```

## Modules

This project leverages a set of foundational modules specifically designed for [cloud deployment](../cloud/modules/).

### EC2

**aws_ec2_kind** - This module facilitates the deployment of an EC2 instance with a Kubernetes cluster using the kind tool. The deployment process relies on a customized image that includes the necessary components for running kind.

Please note that the customized image is currently available in the following regions: **eu-west-1** and **eu-west-3**. Therefore, ensure that you deploy this module in one of these regions to leverage the customized image effectively.

**Module variables**

| Variable  | Default Value | Comment |
|:-------------:|:-------------:|:------------|
| region      |   eu-west-3   | AWS EC2 deployment region |


### Elastic Cloud

**ec_deployment** - This module facilitates the deployment of Elastic Cloud instance.


## Execution

To execute the full project, encompassing the deployment of an EC2 instance, setting up a Kubernetes cluster using kind, and deploying Elastic Cloud, follow the steps outlined below

- Initiate the project

```bash
cd test-environments
terraform init
```

- Deploy test environment

```bash
terraform apply --auto-approve
```

For development purposes, it is possible to deploy each module separately, allowing for focused and independent development and testing. Each module within the project represents a specific component or functionality and can be deployed individually to streamline the development process.

Below are examples demonstrating how to execute individual modules separately:

- EC2 + Kind Kubernetes

```bash
terraform apply --auto-approve -target "module.aws_ec2_kind"
```

- EC Deployment

```bash
terraform apply --auto-approve -target "module.ec_deployment"
```

- EKS Deployment

```bash
terraform apply --auto-approve -target "module.eks"
```