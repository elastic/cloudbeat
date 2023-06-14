# Cloud Deployment

**Motivation**
Provide an easy and deterministic way to set up latest cloud environment, so it can be monitored and used properly.

This guide deploys both an Elastic cloud environment, an AWS EKS cluster, and an Elastic agent on that cluster. To only deploy specific resources, check out the examples section.

**Prerequisite**

Follow the [prerequisites](/README.md#prerequisites) chapter of our main README.

**How To Create an Environment**

1. Create an [API token](https://cloud.elastic.co/deployment-features/keys) from your cloud console account.

   1.1 use the token `export TF_VAR_ec_api_key={TOKEN}`

2. In case you want to deploy a specific stack version, set the `TF_VAR_stack_version` variable to the desired version.

   for `SNAPSHOT` version make sure to also set the region properly.

   ```bash
   export TF_VAR_stack_version=8.7.2-SNAPSHOT
   export TF_VAR_agent_docker_image_override=docker.elastic.co/beats/elastic-agent:8.7.2-SNAPSHOT
   export TF_VAR_ess_region=gcp-us-west2
   ```

   Note: if instead of using environment variables you want to use the `-var` flag, make sure to pass that same variable in all stages of the deployment.

3. To create an EKS cluster and the Elastic cloud environment from the latest version (the latest version is varying in cloud/regions combinations) run:
   ```bash
   cd deploy/cloud
   terraform init
   terraform apply --auto-approve -target "module.ec_deployment" -target "null_resource.rules" -target "null_resource.store_local_dashboard" -target "module.eks"
   ```
4. Create EC2 instance to run Cloudbeat on vanilla cluster (KSPM) and to run CSPM's Cloudbeat agent (from version >=8.7.0)
   ```bash
   terraform apply --auto-approve -target "module.aws_ec2_with_agent"
   ```
   (Note it may take more than 20 minutes to create all the resources)
5. To deploy nginx ingress controller, and ebs csi driver run:
   ```bash
   terraform apply --auto-approve -target "module.apps"
   ```
6. To create an agent policy and IAM role for EKS, run:
   ```bash
   terraform apply --auto-approve -target "module.api" -target "module.iam_eks_role"
   ```
7. To deploy the agent on EKS run:
   ```bash
   terraform apply --auto-approve
   ```
8. Run the following command to retrieve the access credentials for your EKS cluster and configure kubectl.
   ```bash
   aws eks --region $(terraform output -raw eks_region) update-kubeconfig \
       --name $(terraform output -raw eks_cluster_name)
   ```
   To connect to the environment use the console UI or see the details how to connect to the environment, using:
   ```bash
   terraform output -json
   ```

## Modules

We have multiple modules that allows us to deploy different resources based on the intention.

### Elastic Stack

### EKS

### EC2

**Prerequisite: elastic-stack is deployed**
When `-target=module.aws_ec2_with_agent` passed an ec2 instance will be created.
On this instance an agent will be installed to run KSPM integration on vanilla Kubernetes cluster.
To connect to the instance use the generated private key.
See the ssh command `terraform output -raw cloudbeat_ssh_cmd`

**Delete environment:**

```bash
terraform destroy --auto-approve
```

**Next Steps**

- Enable rules add slack webhook to connector

# Examples

## Specific version

To create an environment with specific version use

```bash
terraform apply --auto-approve -var="stack_version=8.7.2"
```

When working with non-production versions it is required to also update the deployment regions.
For example, to deploy `8.7.2-SNAPSHOT` use

```bash
terraform apply --auto-approve \
   -var="stack_version=8.7.2-SNAPSHOT" \
   -var="agent_docker_image_override=docker.elastic.co/beats/elastic-agent:8.7.2-SNAPSHOT" \
   -var="ess_region=gcp-us-west2"
```

## Named environment

To give your environment a name in advance use:

`terraform apply --auto-approve -var="deployment_name=john-8-8-0bc1-30Apr"`

## Deploy specific resources

To deploy specific resources use the `-target` flag.

### Deploy only Elastic Cloud with no EKS cluster or Dashboard

`terraform apply --auto-approve -target "module.ec_deployment"`

### Deploy only Dashboard on an existing Elastic Cloud deployment

`terraform apply --auto-approve -target "null_resource.rules" -target "null_resource.store_local_dashboard"`

### Deploy only EKS cluster

`terraform apply --auto-approve -target "module.eks"`

### Deploy only EC2 instance

**Prerequisite: elastic-stack is running**
`terraform apply --auto-approve -target "module.aws_ec2_with_agent"`
