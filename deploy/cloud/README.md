# Cloud Deployment

**Motivation**
Provide an easy and deterministic way to setup latest cloud environment so it can be monitored and used properly.

**Prerequisite**
* [terraform](https://www.terraform.io/)

**How To**
Create environment
1. Create an [API token](https://cloud.elastic.co/deployment-features/keys) from your cloud console account.
    1.1 use the token `export EC_API_KEY={TOKEN}`
2. run `cd deploy/cloud`
3. run `terraform init`
4. run `terraform apply --auto-approve` to create the environment from the latest version (the latest version is vary in cloud/regions combinations).
To connect to the environment use the console ui or see the details how to connect to the environment, use `terraform output -json`

Delete environment
1. `terraform destroy --auto-approve`

**Next Steps**
* [Setup](https://github.com/elastic/security-team/blob/main/docs/cloud-security-posture-team/onboarding/deploy-agent-cloudbeat-on-eks.mdx) EKS cluster
* Setup Vanila cluster
* Enable rules add slack webhook to connctor

# Examples

## Spesific version
To create an environment with spesific version use `terraform apply --auto-approve -var="stack_version=8.5.1"`
When working with non production versions it is required to also update the deployment regions.
For example, to deploy `8.6.0-SNAPSHOT` use `terraform apply --auto-approve -var="stack_version=8.6.0-SNAPSHOT" -var="ess_region=gcp-us-west2"`

## Named environment
To give your environment a different prefix in the name use `terraform apply --auto-approve -var="deployment_name_prefix=elastic-deployment"`
