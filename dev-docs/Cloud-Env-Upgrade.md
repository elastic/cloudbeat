# Cloud Environment Upgrade Testing

The [`Test Upgrade Environment`](https://github.com/elastic/cloudbeat/actions/workflows/upgrade-environment.yml) GitHub action automates the process of deploying a fully-featured cloud environment, pre-configured with all integrations (KSPM, CSPM, and D4C).
It also facilitates the upgrade of the environment to a new version of the ELK stack and all installed agents, while also performing checks for findings retrieval. For example, if the target ELK version is 8.12.0 and the base version was not selected, the workflow will automatically calculate the previously released version (e.g., 8.11.3), install that version, and then proceed to upgrade to the specified target version (8.12.0). Essentially, this workflow is designed to test the upgrade feature on upcoming versions that are currently in development or will be release candidates (BC).


## Overview of the Upgrade Process

The upgrade process comprises the following main steps:

1. Install the released version, including all integrations (CSPM/KSPM), and deploy their agents.
2. Upgrade the ELK stack version.
3. Upgrade CSPM/KSPM integration versions:
   - If the integration has a `preview` version, the workflow will execute a script to update the integration to the latest `preview` version.
   - If the latest version is released (no `preview` suffix), the integration upgrade will be automatically performed after the stack upgrade.
4. Upgrade KSPM agents by reapplying Kubernetes manifests with the latest image versions.
5. Upgrade Linux-type agents (CSPM/CNVM) by using the Fleet upgrade API.

## How to Run the Workflow

Follow these steps to run the workflow:

1. Go to [`Actions > Test Upgrade Environment`](https://github.com/elastic/cloudbeat/actions/workflows/upgrade-environment.yml).

   ![Navigate to Actions](https://github.com/elastic/cloudbeat/assets/99176494/2686668f-7be6-4b55-a37b-e37426c1a0e1)

2. Click the `Run workflow` button.

   ![Run Workflow](https://github.com/elastic/cloudbeat/assets/99176494/902efe40-ed1b-4175-92a6-504439eb9e3d)

3. Complete the required parameters:

    - **`deployment_name`**: Name your environment (Allowed characters: a-z0-9 and `-`). For
      instance: `john-8-11-0-nov1`.

    - **`target-elk-stack-version`**: Specify the target version for the Elastic Cloud stack upgrade. This version represents the goal to which the workflow will upgrade the stack. Check the available versions [here](https://artifacts-staging.elastic.co/dra-info/index.html).
      For BC, enter version with additions/commit sha, e.g. `8.12.0-61156bc6`.
      For SNAPSHOT, enter the full version, e.g. `8.13.0-SNAPSHOT`.

   ![Required Parameters](https://github.com/elastic/cloudbeat/assets/99176494/9475f553-70c9-4dd7-8330-260bbd704df8)

4. Optionally, modify other parameters if required:
    - **`base-elk-stack-version`** (**optional**): Use this if you're planning to upgrade from a specific released version.
    - **`docker-image-override`** (**optional**): Use this to replace the default Docker image for build candidate (BC) or
      SNAPSHOT versions.
      Provide the full image path. Leave this field blank for snapshot versions. Follow this format for the image
      path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.11.0-cb971279`. If you're not sure where to get this
      image path from, look for message like [this](https://elastic.slack.com/archives/C0JFN9HJL/p1698263174847419) in
      #mission-control channel, you can see it specify the stack version and the BC commit sha in the first line,
      e.g. `elastic / unified-release - staging # 8.11 - 10 - 8.9.0-cb971279`. Now just copy it
      and replace it the image path: `docker.elastic.co/cloud-release/elastic-agent-cloud:8.11.0-cb971279`.
    - **`kibana_ref`** (**optional**): Specifies the Kibana branch, tag, or commit SHA to check out for the UI sanity tests, which will be executed after the environment is upgraded. This should correspond to the version of the `target-elk-stack-version` provisioned by this workflow. For the current version in development, use Kibana's `main` branch. Default: `main`. Examples of different inputs:
      - Specifying Branch: `main`
      - Specifying Tag: `v8.13.4`
      - Specifying Commit SHA: `c776cf650e962f04330789a9f113bd4bbd6d7c61`

   ![Optional Parameters](https://github.com/user-attachments/assets/a0e1b61d-ea5a-4166-b1fa-23291e094317)


## Tracking Workflow Execution

Tracking workflow execution follows the same steps as defined in the [Create Environment](./Cloud-Env-Testing.md#tracking-workflow-execution)

## Logging into the Environment

Logging into the environment can be done following the steps detailed in the [Create Environment](./Cloud-Env-Testing.md#logging-into-the-environment)

## Cleanup Procedure

The cleanup procedure is also described in the [Create Environment](./Cloud-Env-Testing.md#cleanup-procedure)
