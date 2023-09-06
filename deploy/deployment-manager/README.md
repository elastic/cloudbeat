## Elastic Agent Deployment manager template

### What it does
This template creates a compute instance and attach a service account with a custom role with the relevant permissions to run the CIS GCP integration.
The compute engine instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
*Prerequisites:*
1. You have an elastic stack deployed in the cloud that includes Kibana, elasticsearch and fleet-server (check https://github.com/elastic/cloudbeat/blob/main/dev-docs/ELK-Deployment.md to deploy your own stack)
2. You have been authenticated to GCP by running `gcloud auth login` and configured to work with our dev account `elastic-security-test`.

*Steps:*
1. Install the GCP CSPM integration on a new agent policy, you might have to check the "Display beta integrations" checkbox.
2. After you installed the integration, deploy a new agent, you should keep the fleet URL and the enrollment token.
3. Run `DEPLOYMENT_NAME=<NAME> FLEET_URL=<URL> ENROLLMENT_TOKEN=<TOKEN> ELASTIC_ARTIFACT_SERVER=<ARTIFACT_SERVER> STACK_VERSION=<VERSION> ZONE=<ZONE> ALLOW_SSH=<true|false> just deploy-dm` to create a new deployment with an elastic-agent that will automatically enroll to your fleet.
```
DEPLOYMENT_NAME="<Unique stack name>" # john-qa-bc2-8-9-0-May28
FLEET_URL="<Elastic Agent Fleet URL>"
ENROLLMENT_TOKEN="<Elastic Agent Enrollment Token>"
ELASTIC_ARTIFACT_SERVER="https://artifacts.elastic.co/downloads/beats/elastic-agent" # Replace artifact URL with a pre-release version (BC or snapshot)
ELASTIC_AGENT_VERSION="<Elastic Agent Version>" # e.g: 8.8.0 | 8.8.0-SNAPSHOT
ZONE="<GCP Zone>" # e.g: us-central1-a
ALLOW_SSH=false # Set to true to allow SSH connections to the deployed instance
```
4. For running the integration at the **Organization level**, run the same command mentioned in step 3 but with another environment variable `ORG_ID=<ORG_ID>`.
5. For deleting the deployment, run `just delete-dm <DEPLOYMENT_NAME>`.

*Debugging:*
1. Deployments creation may take a few minutes, to see the progress, find your deployment on https://console.cloud.google.com/dm/deployments/ and click on it.
2. If the deployment was created successfully but elastic-agent didn't enroll to your fleet, try to ssh into the EC2 by running `gcloud compute ssh elastic-agent --zone <ZONE>` and then get the startup logs by running `sudo journalctl -u google-startup-scripts.service`.
3. Re-running the startup script can be done by running `sudo google_metadata_script_runner startup`.
