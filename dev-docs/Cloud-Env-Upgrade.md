# Cloud Environment Upgrade Testing

The [`Test Upgrade Environment`](https://github.com/elastic/cloudbeat/actions/workflows/upgrade-environment.yml) GitHub action automates the process of deploying a fully-featured cloud environment, pre-configured with all integrations (KSPM, CSPM, and D4C).
It also facilitates the upgrade of the environment to a new version of the ELK stack and all installed agents, along with performing findings retrieval checks.


## How to Run the Workflow

Follow these steps to run the workflow:

1. Go to [`Actions > Test Upgrade Environment`](https://github.com/elastic/cloudbeat/actions/workflows/upgrade-environment.yml).

   ![Navigate to Actions](https://github.com/elastic/cloudbeat/assets/99176494/2686668f-7be6-4b55-a37b-e37426c1a0e1)

2. Click the `Run workflow` button.

   ![Run Workflow](https://github.com/elastic/cloudbeat/assets/99176494/902efe40-ed1b-4175-92a6-504439eb9e3d)

3. Complete the required parameters:

    - **`deployment_name`**: Name your environment (Allowed characters: a-z0-9 and `-`). For
      instance: `john-8-11-0-nov1`.

    - **`elk-stack-version`**: Specify the version of Elastic Cloud stack, either a SNAPSHOT or a build candidate (BC)
      version. Check the available versions [here](https://artifacts-staging.elastic.co/dra-info/index.html).
      For BC, enter only the version without additions/commit sha, e.g. `8.11.0`.
      For SNAPSHOT, enter the full version, e.g. `8.12.0-SNAPSHOT`.

   ![Required Parameters](https://github.com/elastic/cloudbeat/assets/99176494/a50141d7-7554-4761-a737-e0f23f0b0492)

4. Optionally, modify other parameters if required:

    - **`docker-image-override`** (**optional**): Use this to replace the default Docker image for build candidate (BC) or
      SNAPSHOT versions.
      Provide the full image path. Leave this field blank for snapshot versions. Follow this format for the image
      path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.11.0-cb971279`. If you're not sure where to get this
      image path from, look for message like [this](https://elastic.slack.com/archives/C0JFN9HJL/p1698263174847419) in
      #mission-control channel, you can see it specify the stack version and the BC commit sha in the first line,
      e.g. `elastic / unified-release - staging # 8.11 - 10 - 8.9.0-cb971279`. Now just copy it
      and replace it the image path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.11.0-cb971279`.

   ![Optional Parameters](https://github.com/elastic/cloudbeat/assets/99176494/5b7f15bd-6f56-4eb0-b7d6-fc6a7656ffb0)

## Tracking Workflow Execution

Tracking workflow execution follows the same steps as defined in the [Create Environment](./Cloud-Env-Testing.md#tracking-workflow-execution)

## Logging into the Environment

Logging into the environment can be done following the steps detailed in the [Create Environment](./Cloud-Env-Testing.md#logging-into-the-environment)

## Cleanup Procedure

The cleanup procedure is also described in the [Create Environment](./Cloud-Env-Testing.md#cleanup-procedure)
