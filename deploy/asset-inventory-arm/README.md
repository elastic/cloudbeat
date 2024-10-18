## ARM deployment for developers

The [`generate_dev_template.py`](./generate_dev_template.py) script generates an ARM template for deploying the Elastic
Agent with SSH access enabled to the VM. This script works both for the single subscription and management group
templates.

Usage:

```text
usage: generate_dev_template.py [-h]
                                [--template-type {single-account,organization-account}]
                                [--output-file OUTPUT_FILE] [--deploy]
                                [--resource-group RESOURCE_GROUP]
                                [--public-ssh-key PUBLIC_SSH_KEY]
                                [--artifact-server ARTIFACT_SERVER]
                                [--elastic-agent-version ELASTIC_AGENT_VERSION]
                                [--fleet-url FLEET_URL]
                                [--enrollment-token ENROLLMENT_TOKEN]

Deploy Azure resources for a single account

options:
  -h, --help            show this help message and exit
  --template-type {single-account,organization-account}
                        The type of template to use
  --output-file OUTPUT_FILE
                        The output file to write the modified template to
  --deploy              Perform deployment
  --resource-group RESOURCE_GROUP
                        The resource group to deploy to
  --public-ssh-key PUBLIC_SSH_KEY
                        SSH public key to use for the VMs
  --artifact-server ARTIFACT_SERVER
                        The URL of the artifact server
  --elastic-agent-version ELASTIC_AGENT_VERSION
                        The version of elastic-agent to install
  --fleet-url FLEET_URL
                        The fleet URL of elastic-agent
  --enrollment-token ENROLLMENT_TOKEN
                        The enrollment token of elastic-agent
```

Arguments are also read from the `dev-flags.conf` file in the same directory as the script. Write the arguments in the
file as you would pass them to the script. Notice that you need to properly quote arguments. Example:

```text
--artifact-server https://snapshots.elastic.co/8.12.0-t9e0i58r/downloads/beats/elastic-agent
--elastic-agent-version 8.12.0-SNAPSHOT
--fleet-url <fleet url>
--enrollment-token <enrollment token>
--public-ssh-key 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC3e38/Q26WUsyUVb4D7N1McL9QbrcamMfZw23+txivvP13QXzIEyvMjsqUpX0kqjg+C4OD7osfZ+wlVI3QFkomjDjjPMx/FYGUGk5ZKvKh9vXyxN2brYZq8C24lWQSpZbmvNF4+FueFx1eo6wMllLzmzzQ60LpeBhNhRiDPiLQKBotDn1mD6zymnhSANpS/+rWX5HVguSQgtEZP4vvxpKVxEM8hnT8V0PvWFfuNQpTf7zVpZtFvGTLoosvvGbQ27wiufHdF8vv9mF5cXhy02N4IaREcJEMu5wmQaD7zUcJ67aN4v7FTwkA6D3sppb7cJolUJJiOWh4kt7K03BEBYIM9g88lhHDFxwpUvMNWhwp/RHnu8/Ic3HL623W5EDcXxsjH1gsIpXtNuSaUP6G+c2k1zvmST7Oom6EXLT47hv9MXWcS7zY1YZtqVlboZiBRH5MfqwRPFHl6r04yqq1vithW/LeBweH8/q4iWaVYABda0Zmq8qFKKu/5VZStqbOt5wa0bIZrMn+dU6NUHlP6gOuM1yb7kbR2Y/x7AnHvNZ8YtcXDmoMjX93/7A+4Dr3qZd0FKtVoYqUspg0jOGH/Kj3sswp7oM98yJz5F/3/7VwSdzO/DzSGr9Of9BLCQHfcS6qJUZjsErPDqc0T7v7c+Dsz73t5zYq8uYovtUt6m3Anw== user@hostname'
--deploy
```

Executing the deployment with `--deploy` requires the `az` CLI to be installed and logged in to the correct
subscription.

The script is included the pre-commit pipelines so new dev templates will be generated each time a change is made to the
source templates.
