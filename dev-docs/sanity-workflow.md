# Sanity Tests

This GitHub workflow deploys the environment, saves state and data, performs sanity checks, and provides an option to destroy the infrastructure.

## Running the Workflow

To run the workflow, perform the following steps:

1. Click on the "Actions" tab in Cloudbeat repository.

![image](https://github.com/elastic/cloudbeat/assets/99176494/2686668f-7be6-4b55-a37b-e37426c1a0e1)

2. Select the "Sanity Tests" workflow.

![image](https://github.com/elastic/cloudbeat/assets/99176494/f2e8ce8f-11f5-483d-b067-b24db3f58114)

3. Click on the "Run workflow" button.

![image](https://github.com/elastic/cloudbeat/assets/99176494/115fdd53-cff7-406a-bc3d-d65d5199389f)

4. Fill in the required inputs: `deployment_name`, `ec-api-key`, `elk-stack-version`, `ess-region`.

![image](https://github.com/elastic/cloudbeat/assets/99176494/06d8144d-13cc-4e13-92fc-19f52ce8206b)

5. Optionally, adjust the other input values as needed.

![image](https://github.com/elastic/cloudbeat/assets/99176494/bac5004d-7cbc-4a34-8127-3acd11acc90e)

6. Click on the "Run workflow

![image](https://github.com/elastic/cloudbeat/assets/99176494/5e5131ba-264e-4444-8879-aa612d5de778)


To track the execution of the Sanity flow, follow these steps:

1. Click on the "Sanity tests" to access its details.

![image](https://github.com/elastic/cloudbeat/assets/99176494/abe8182d-4229-41bd-8604-ed5202d23574)


2. Click on "Deploy"

![image](https://github.com/elastic/cloudbeat/assets/99176494/230743cf-02ff-40cb-9069-d747b460824c)

3. Once the flow execution is complete, click on the "Summary" button to get the summary report.

![image](https://github.com/elastic/cloudbeat/assets/99176494/7751d919-1605-4d07-9cfd-c98336051e3d)

4. Review Summary details

![image](https://github.com/elastic/cloudbeat/assets/99176494/1b41fba0-0ee5-4d37-b2f8-cdd6f632eadc)


## Inputs

- `deployment_name` (required): Name your environment (Only a-zA-Z0-9 and `-`). For example: `john-8-7-2-June01`.
- `ec-api-key` (required): Elastic Cloud API KEY.
- `elk-stack-version` (required): Elastic Cloud stack SNAPSHOT or BC version. Default: `8.7.2-SNAPSHOT`.
- `ess-region` (required): Elastic Cloud deployment region. Default: `gcp-us-west2`.
- `docker-image-override` (optional): Provide the full Docker image path to override the default image (BC versions only).
- `cleanup-env` (optional): Boolean value to indicate if resources should be cleaned up after provision. Default: `false`.

## Environment Variables

The following environment variables are required:

- `AWS_ACCESS_KEY_ID`: Your AWS access key ID.
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key.
- `AWS_REGION`: The AWS region (e.g., `eu-west-1`).
- `TF_VAR_ec_api_key`: Your Elastic Cloud API key.
- `WORKING_DIR`: The working directory for the deployment.
- `FLEET_API_DIR`: The directory for the Fleet API.

## Cleanup

If you want to destroy the provisioned infrastructure, set the `cleanup-env` input to `true` when running the workflow. The cleanup step will be executed at the end.

Note: Destroying the environment is irreversible, so use it with caution.