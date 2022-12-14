# Cloud Deployment

**Motivation**
Provide an easy and deterministic way to setup latest cloud environment so it can be monitored and used properly.

**Prerequisite**
* [Terraform](https://developer.hashicorp.com/terraform/downloads)
* the AWS CLI, [installed](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) and [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
* [AWS IAM Authenticator](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html)
* the [Kubernetes CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/), also known as `kubectl`
* [jq](https://stedolan.github.io/jq/download/)


**How To**
Create environment
1. Create an [API token](https://cloud.elastic.co/deployment-features/keys) from your cloud console account.

    1.1 use the token `export EC_API_KEY={TOKEN}`

2. run `cd deploy/cloud`
3. run `terraform init`
4. run `terraform apply --auto-approve` to create the environment from the latest version (the latest version is vary in cloud/regions combinations).
5. Run the following command to retrieve the access credentials for your EKS cluster and configure kubectl.
```bash
aws eks --region $(terraform output -raw eks_region) update-kubeconfig \
    --name $(terraform output -raw eks_cluster_name)
```
To connect to the environment use the console ui or see the details how to connect to the environment.

Delete environment
1. `terraform destroy --auto-approve`

**Next Steps**
* [Setup](https://github.com/elastic/security-team/blob/main/docs/cloud-security-posture-team/onboarding/deploy-agent-cloudbeat-on-eks.mdx) EKS cluster
* Setup Vanilla cluster
* Enable rules add slack webhook to connector

# Examples

## Specific version
To create an environment with specific version use `terraform apply --auto-approve -var="stack_version=8.5.1"`

## Named environment
To give your environment a different prefix in the name use `terraform apply --auto-approve -var="deployment_name_prefix=elastic-deployment"`
