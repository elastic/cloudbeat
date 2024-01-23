# Cloud Environment Testing

The [`Create Environment`](https://github.com/elastic/cloudbeat/actions/workflows/test-environment.yml) GitHub action
deploys a full-featured cloud environment, pre-configured with all our integrations (KSPM, CSPM and D4C).
It also includes features for running sanity testing and automated deletion.

## How to Run the Workflow

Follow these steps to run the workflow:

1. Go to [`Actions > Create Environment`](https://github.com/elastic/cloudbeat/actions/workflows/test-environment.yml).

   ![Navigate to Actions](https://github.com/elastic/cloudbeat/assets/99176494/2686668f-7be6-4b55-a37b-e37426c1a0e1)

2. Click the `Run workflow` button.

   ![Run Workflow](https://github.com/elastic/cloudbeat/assets/99176494/115fdd53-cff7-406a-bc3d-d65d5199389f)

3. Complete the required parameters:

    - **`deployment_name`**: Name your environment (Allowed characters: a-zA-Z0-9 and `-`). For
      instance: `john-8-7-2-June01`.

    - **`elk-stack-version`**: Specify the version of Elastic Cloud stack, either a SNAPSHOT or a build candidate (BC)
      version. Check the available versions [here](https://artifacts-staging.elastic.co/dra-info/index.html).
      For BC, enter version with additions/commit sha, e.g. `8.12.0-61156bc6`.
      For SNAPSHOT, enter the full version, e.g. `8.13.0-SNAPSHOT`.

    - **`ess-region`**: Indicate the Elastic Cloud deployment region. The default value is `gcp-us-west2`, which
      supports
      snapshot and build candidate (BC) versions. Specify a different region only if necessary.

   ![Required Parameters](https://github.com/oren-zohar/cloudbeat/assets/85433724/6159129e-6d4d-46b1-97a1-f0d3859500fd)

4. Optionally, modify other parameters if required:

    - **`docker-image-override`** (**optional**): Use this to replace the default agent Docker image for build candidate (BC) or
      SNAPSHOT versions.
      Provide the full image path. Leave this field blank for snapshot versions. Follow this format for the image
      path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.8.1-9ac7eb02`. If you're not sure where to get this
      image path from, look for message like [this](https://elastic.slack.com/archives/C0JFN9HJL/p1689227472876399) in
      #mission-control channel, you can see it specify the stack version and the BC commit sha in the first line,
      e.g. `elastic / unified-release - staging # 8.9 - 11 - 8.9.0-c6bb8f7a Success after 4 hr 58 min`. Now just copy it
      and replace it the image path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.9.0-c6bb8f7a`.

    - **`run-sanity-tests`** (**optional**): Set to `true` to run sanity tests after the environment is set up. Default: `false`

    - **`cleanup-env`** (**optional**): Set to `true` if you want the resources to automatically be cleaned up after
      provisioning - useful if you don't want to test the env manually after deployment.
      Default: `false`.

    - **`ec-api-key`** (**optional**): By default, all the new environments will be created in our EC Cloud Security organization.
      If you want to create the environment on your personal org (`@elastic.co`) you can enter
      your private [Elastic Cloud](https://cloud.elastic.co/home) API key. Follow the
      [Cloud API Keys](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html) documentation for
      step-by-step instructions on generating the token.

   ![Optional Parameters](https://github.com/oren-zohar/cloudbeat/assets/85433724/17933589-ee0e-4181-a244-f501f54bda6c)

5. Click the `Run workflow` button to start.

   ![Run Workflow](https://github.com/oren-zohar/cloudbeat/assets/85433724/7b05bf58-cc0b-4ec9-8e49-55d117673df8)


## Tracking Workflow Execution

Monitor the progress of the workflow execution as follows:

1. Click `Create Environment` for details.

   ![Create Environment](https://github.com/elastic/cloudbeat/assets/99176494/abe8182d-4229-41bd-8604-ed5202d23574)

2. Click `Deploy`.

   ![Deploy](https://github.com/elastic/cloudbeat/assets/99176494/230743cf-02ff-40cb-9069-d747b460824c)

3. Once the workflow execution finishes, click the `Summary` button to view the summary report.

   ![Summary Report](https://github.com/elastic/cloudbeat/assets/99176494/7751d919-1605-4d07-9cfd-c98336051e3d)

4. Review the details in the Summary.

   ![Summary Details](https://github.com/elastic/cloudbeat/assets/99176494/1b41fba0-0ee5-4d37-b2f8-cdd6f632eadc)

## Logging into the Environment

Follow these steps to log in to the created environment:

1. Click the `kibana` link.

   ![Kibana Link](https://github.com/elastic/cloudbeat/assets/99176494/500351cf-6029-4bd5-bc6f-e6e046fbb73d)

2. Select `Login with Elastic Cloud`.

   ![Login](https://github.com/elastic/cloudbeat/assets/99176494/c3c1521e-e997-43ce-af76-b00aa0fa353a)

3. Choose the `Google` authentication method.

   ![Google Authentication](https://github.com/elastic/cloudbeat/assets/99176494/f5209ed8-3bd7-420e-a3d1-cffb4c3711c9)

4. In the Elastic Cloud dashboard, click `Open` next to the created environment.

![Open Environment](https://github.com/elastic/cloudbeat/assets/99176494/b2bcf5f3-d463-4d2c-8073-8ef9183c9ada)

## Access AWS EKS Cluster

Follow these steps to connect to your Amazon Elastic Kubernetes Service (EKS) cluster:

1. **Assume Role for Access**:

   Before connecting to the EKS cluster, you need to assume a role that provides the necessary permissions.
   Replace `<your-session-name>` with a meaningful session name and run the following command to assume the role:

   ```bash
   export $(printf "AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s AWS_SESSION_TOKEN=%s"  $(aws sts assume-role --role-arn arn:aws:iam::704479110758:role/Developer_eks --role-session-name <your-session-name> --query "Credentials.[AccessKeyId,SecretAccessKey,SessionToken]" --output text))
   ```

   This command sets temporary AWS credentials that grant you access to your EKS cluster.

2. **Update Kubeconfig**:

   To configure kubectl to communicate with your EKS cluster, replace `<cluster_name>` with your EKS cluster's name and run the following command:

   ```aws eks update-kubeconfig --region eu-west-1 --name <cluster_name>```

   This command updates your ~/.kube/config file with the necessary cluster configuration.

3. **Check Connectivity**:

   To verify your connectivity to the EKS cluster, run the following kubectl command:

   ```kubectl get po -n kube-system```

   This command should list the pods in the kube-system namespace, confirming that you have successfully connected to your EKS cluster.


## Cleanup Procedure

If you wish to automatically delete the environment after the tests finish, set the `cleanup-env` input to `true`.

In addition to the automatic cleanup, you can manually delete environments using the [Destroy Environment](https://github.com/elastic/cloudbeat/actions/workflows/destroy-environment.yml) workflow or by directly executing the `delete-cloud-env` command.

### Destroy Environment Workflow

The **Destroy Environment** GitHub workflow automates the process of cleaning up environments. When activated, it automatically performs the cleanup of environments, ensuring that all associated resources are correctly removed.

#### How to Run the Flow

Follow these steps to run the workflow:

1. Go to [`Actions > Destroy Environment`](https://github.com/elastic/cloudbeat/actions/workflows/destroy-environment.yml)

   ![Destroy Environment](https://github.com/gurevichdmitry/cloudbeat/assets/99176494/505d6553-7780-4450-83e9-3617f64442ad)

2. Click the `Run workflow` button.

   ![Run Workflow](https://github.com/gurevichdmitry/cloudbeat/assets/99176494/8965311c-eeac-492f-a564-a57c46854a3a)

3. Complete the required input fields:

    - `prefix` (required): The prefix used to identify the environments to be deleted.

   <img width="411" alt="Enter Inputs" src="https://github.com/elastic/cloudbeat/assets/99176494/04973b00-5411-4ace-ab3a-534371877c91">

4. Optionally, modify other input value if required:

    - `ignore-prefix` (optional): The prefix used to identify environments that should be excluded from deletion.
    - `ec-api-key` (required): Use your own [Elastic Cloud](https://cloud.elastic.co/home) API key if you want to delete environments from your Elastic Cloud account.

   <img width="411" alt="Optional Inputs" src="https://github.com/elastic/cloudbeat/assets/99176494/aa89ad4e-fd32-461d-ab2d-3fee28094a9d">

5. Click the `Run workflow` button to start.

   ![Run Workflow](https://github.com/gurevichdmitry/cloudbeat/assets/99176494/64b554d5-70f0-4cf3-b2b9-8f8429d1fc07)

### Manual Environment Deletion

In addition to the automatic cleanup, you can manually delete environments using:

```bash
just delete-cloud-env <prefix> <ignore-prefix> <interactive>
```

This script deletes all environments that match a given prefix, and ignores environments that match a given ignore
prefix.

Before running the script, ensure that:

- The AWS CLI is installed and configured.
- The Terraform CLI is installed and configured.
- The `TF_VAR_ec_api_key` environment variable is set.

**Note**: The script will ask for confirmation before deleting each environment, unless you set the `interactive` flag
to `false`.
