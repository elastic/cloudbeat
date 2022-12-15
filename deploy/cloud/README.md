# Cloud Deployment

**Motivation**
Provide an easy and deterministic way to set up latest cloud environment, so it can be monitored and used properly.

This guide deploys both an Elastic cloud environment, and an AWS EKS cluster. To only deploy specific resources, check out the examples section.

**Prerequisite**
* [Terraform](https://developer.hashicorp.com/terraform/downloads)
* the AWS CLI, [installed](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) and [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
* [AWS IAM Authenticator](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html)
* the [Kubernetes CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/), also known as `kubectl`


**How To**
Create environment
1. Create an [API token](https://cloud.elastic.co/deployment-features/keys) from your cloud console account.

    1.1 use the token `export TF_VAR_ec_api_key={TOKEN}`

2. run `cd deploy/cloud`
3. run `terraform init`
4. run `terraform apply --auto-approve` to create the environment from the latest version (the latest version is varying in cloud/regions combinations).
5. Run the following command to retrieve the access credentials for your EKS cluster and configure kubectl.
```bash
aws eks --region $(terraform output -raw eks_region) update-kubeconfig \
    --name $(terraform output -raw eks_cluster_name)
```

To connect to the environment use the console ui or see the details how to connect to the environment, use `terraform output -json`

Delete environment
1. `terraform destroy --auto-approve`

**Next Steps**
* [Setup](https://github.com/elastic/security-team/blob/main/docs/cloud-security-posture-team/onboarding/deploy-agent-cloudbeat-on-eks.mdx) EKS cluster
* Setup Self-Managed cluster
* Enable rules add slack webhook to connector

# Examples

## Specific version
To create an environment with specific version use

`terraform apply --auto-approve -var="stack_version=8.5.1"`
When working with non production versions it is required to also update the deployment regions.
For example, to deploy `8.6.0-SNAPSHOT` use

`terraform apply --auto-approve -var="stack_version=8.6.0-SNAPSHOT" -var="ess_region=gcp-us-west2"`

## Named environment
To give your environment a different prefix in the name use

`terraform apply --auto-approve -var="deployment_name_prefix=elastic-deployment"`

## Deploy specific resources
To deploy specific resources use the `-target` flag.

### Deploy only Elastic Cloud with no EKS cluster or Dashboard

`terraform apply --auto-approve -target "module.ec_deployment"`

### Deploy only Dashboard on an existing Elastic Cloud deployment

`terraform apply --auto-approve -target "null_resource.rules" -target "null_resource.store_local_dashboard"`

### Deploy only EKS cluster

`terraform apply --auto-approve -target "module.eks"`
