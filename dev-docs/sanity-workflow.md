# Sanity Tests

This GitHub workflow deploys the environment, saves state and data, performs sanity checks, and provides an option to destroy the infrastructure.

## Running the Workflow

To run the workflow, perform the following steps:

1. Click on the `Actions` tab in [Cloudbeat](https://github.com/elastic/cloudbeat) repository.

![image](https://github.com/elastic/cloudbeat/assets/99176494/2686668f-7be6-4b55-a37b-e37426c1a0e1)

1. Select the `Sanity Tests` workflow. If the workflow is not visible, click on `Show more workflows...` link

![image](https://github.com/elastic/cloudbeat/assets/99176494/f2e8ce8f-11f5-483d-b067-b24db3f58114)

3. Click on the `Run workflow` button.

![image](https://github.com/elastic/cloudbeat/assets/99176494/115fdd53-cff7-406a-bc3d-d65d5199389f)

4. Fill in the required inputs: `deployment_name`, `ec-api-key`, `elk-stack-version`, `ess-region`.

- `deployment_name` (required): Name your environment (Only a-zA-Z0-9 and `-`). For example: `john-8-7-2-June01`.

- `ec-api-key` (required): Elastic Cloud API KEY. Follow the [Cloud API Keys](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html) documentation for step-by-step instructions on generating the token.

- `elk-stack-version` (required): The version of Elastic Cloud stack, either a SNAPSHOT or a build candidate (BC) version. The default value is `8.7.2-SNAPSHOT`. You can find the available versions [here](https://artifacts-staging.elastic.co/dra-info/index.html).

- `ess-region` (required): Elastic Cloud deployment region. By default, use the value `gcp-us-west2`, which includes snapshot versions and build candidate (BC) versions. Only specify a different region if there are specific requirements.


![image](https://github.com/elastic/cloudbeat/assets/99176494/06d8144d-13cc-4e13-92fc-19f52ce8206b)

5. Optionally, adjust the other input values as needed.

- `docker-image-override` (optional): To override the default Docker image for build candidate (BC) versions, provide the full image path. For snapshot versions, leave this field empty. The image path should follow this format: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.8.1-9ac7eb02`, where `8.8.1-9ac7eb02` should be replaced with the latest build candidate version.

- `cleanup-env` (optional): Boolean value to indicate if resources should be cleaned up after provision. Default: `false`.

![image](https://github.com/elastic/cloudbeat/assets/99176494/bac5004d-7cbc-4a34-8127-3acd11acc90e)

6. Click on the `Run workflow`

![image](https://github.com/elastic/cloudbeat/assets/99176494/5e5131ba-264e-4444-8879-aa612d5de778)


### To track the execution of the Sanity flow, follow these steps:

1. Click on the `Sanity tests` to access its details.

![image](https://github.com/elastic/cloudbeat/assets/99176494/abe8182d-4229-41bd-8604-ed5202d23574)


2. Click on `Deploy`

![image](https://github.com/elastic/cloudbeat/assets/99176494/230743cf-02ff-40cb-9069-d747b460824c)

3. Once the flow execution is complete, click on the `Summary` button to get the summary report.

![image](https://github.com/elastic/cloudbeat/assets/99176494/7751d919-1605-4d07-9cfd-c98336051e3d)

4. Review Summary details

![image](https://github.com/elastic/cloudbeat/assets/99176494/1b41fba0-0ee5-4d37-b2f8-cdd6f632eadc)


### Environment Login Instructions

To log in to the created environment, please follow these steps:

1. Click on the `kibana` link.

![image](https://github.com/elastic/cloudbeat/assets/99176494/500351cf-6029-4bd5-bc6f-e6e046fbb73d)

2. Select the `Login with Elastic Cloud` option.

![image](https://github.com/elastic/cloudbeat/assets/99176494/c3c1521e-e997-43ce-af76-b00aa0fa353a)

3. Choose the `Google` authentication method.

![image](https://github.com/elastic/cloudbeat/assets/99176494/f5209ed8-3bd7-420e-a3d1-cffb4c3711c9)

4. On the Elastic Cloud dashboard, click on `Open` next to the provisioned environment.

![image](https://github.com/elastic/cloudbeat/assets/99176494/b2bcf5f3-d463-4d2c-8073-8ef9183c9ada)


## Cleanup

If you want to destroy the provisioned infrastructure, set the `cleanup-env` input to `true` when running the workflow. The cleanup step will be executed at the end.
