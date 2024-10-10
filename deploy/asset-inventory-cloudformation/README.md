#TODO(kuba): Update readme


## Elastic Agent EC2 CloudFormation template

### What it does
This CloudFormation template creates a role for elastic-agent and attaches it to a newly created EC2 instance.
The EC2 instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
*Prerequisites:*
1. You have an elastic stack deployed in the cloud that includes Kibana, elasticsearch and fleet-server (check https://github.com/elastic/cloudbeat/blob/main/dev-docs/ELK-Deployment.md to deploy your own stack)
2. You have AWS CLI installed on your laptop and configured to work with our dev account `elastic-security-cloud-security-dev` (in particular, `~/.aws/config` and `~/.aws/credentials` should be set, check https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html for more information)

*Steps:*
1. Install the Vulnerability Management integration on a new agent policy, you might have to check the "Display beta integrations" checkbox.
2. After you installed the integration you can install a new elastic-agent, you should keep the fleet URL and the enrollment token.
3. On cloudbeat repo, create a `deploy/cloudformation/config.env` file of the form:
```
STACK_NAME="<Unique stack name>" # john-qa-bc2-8-9-0-May28
FLEET_URL="<Elastic Agent Fleet URL>"
ENROLLMENT_TOKEN="<Elastic Agent Enrollment Token>"
ELASTIC_ARTIFACT_SERVER="https://artifacts.elastic.co/downloads/beats/elastic-agent" # Replace artifact URL with a pre-release version (BC or snapshot)
ELASTIC_AGENT_VERSION="<Elastic Agent Version>" # e.g: 8.8.0 | 8.8.0-SNAPSHOT
DEPLOYMENT_TYPE="<Type>" # e.g: CNVM | CSPM (default is CNVM)

DEV.ALLOW_SSH=false # Set to true to allow SSH connections to the deployed instance
DEV.KEY_NAME="" # When SSH is allowed, you must provide the key name that will be used to ssh into the EC2
```
4. Run `just deploy-cloudformation` to create a CloudFormation stack with an elastic-agent that will automatically enroll to your fleet.

*Debugging:*
1. CloudFormation stack creation may take a few minutes, to see the progress, find your stack on https://console.aws.amazon.com/cloudformation/ and check the "Event" tab.
2. If the stack was created successfully but elastic-agent didn't enroll to your fleet, try to ssh into the EC2 by running `ssh -i ~/.ssh/<EC2 Key File> ubuntu@<EC2 IP Address>` and then get the initialization logs by `cat /var/log/cloud-init-output.log`.
3. If ssh is not enabled, you can get the system logs from the EC2 instance
![right click on instance -> monitor and troubleshoot -> get sytem logs](https://github.com/orestisfl/cloudbeat/assets/5778622/3f0158ed-ba46-4fe5-b7e9-7a800b3de020)
![system logs window](https://github.com/orestisfl/cloudbeat/assets/5778622/af8105d1-7d6c-47d0-b386-eea124816b53)
You might need to wait a bit until the logs become available but terminating the instance doesn't immediately delete them
