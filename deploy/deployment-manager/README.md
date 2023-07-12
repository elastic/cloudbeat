## Elastic Agent Deployment manager template

### What it does
This template creates a service account to be used by the machine running the elastic-agent.
The compute engine instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
*Prerequisites:*
1. You have an elastic stack deployed in the cloud that includes Kibana, elasticsearch and fleet-server (check https://github.com/elastic/cloudbeat/blob/main/dev-docs/ELK-Deployment.md to deploy your own stack)
2. You have been authenticated to GCP by running `gcloud auth login` and configured to work with our dev account `elastic-security-test`.

*Steps:*
1. Install the GCP CSPM integration on a new agent policy, you might have to check the "Display beta integrations" checkbox.
2. After you installed the integration you can install a new elastic-agent, you should keep the fleet URL and the enrollment token.
3. On cloudbeat repo, create a `deploy/deployment-manager/config.env` file of the form:
```
DEPLOYMENT_NAME="<Unique stack name>" # john-qa-bc2-8-9-0-May28
FLEET_URL="<Elastic Agent Fleet URL>"
ENROLLMENT_TOKEN="<Elastic Agent Enrollment Token>"
ELASTIC_ARTIFACT_SERVER="https://artifacts.elastic.co/downloads/beats/elastic-agent" # Replace artifact URL with a pre-release version (BC or snapshot)
ELASTIC_AGENT_VERSION="<Elastic Agent Version>" # e.g: 8.8.0 | 8.8.0-SNAPSHOT
ZONE="<GCP Zone>" # e.g: us-central1-a
ALLOW_SSH=false # Set to true to allow SSH connections to the deployed instance
```
4. Run `just deploy-deployment-manager` to create a new deployment with an elastic-agent that will automatically enroll to your fleet.

*Debugging:*
1. Deployments creation may take a few minutes, to see the progress, find your deployment on https://console.cloud.google.com/dm/deployments/ and click on it.
2. If the deployment was created successfully but elastic-agent didn't enroll to your fleet, try to ssh into the EC2 by running `gcloud compute ssh elastic-agent --zone <ZONE>` and then get the startup logs by running `sudo journalctl -u google-startup-scripts.service`.
3. Re-running the startup script can be done by running `sudo google_metadata_script_runner startup`.
