## Elastic Agent EC2 CloudFormation template

### What it does
This CloudFormation template creates a role for elastic-agent and attaches it to a newly created EC2 instance.
The EC2 instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
The template can be tested with AWS CLI as follows:
```
<<<<<<< HEAD
aws cloudformation create-stack --stack-name elastic-agent-ec2          \
    --template-body file://deploy/cloudformation/elastic-agent-ec2.yml  \
    --capabilities CAPABILITY_IAM   \
    --parameters                    \
    ParameterKey=ElasticAgentVersion,ParameterValue=elastic-agent-8.6.0-linux-x86_64  \
    ParameterKey=FleetUrl,ParameterValue=<Elastic Agent Fleet URL>                    \
    ParameterKey=EnrollmentToken,ParameterValue=<Elastic Agent Enrollment Token>
=======
STACK_NAME="<Unique stack name>" # john-qa-bc2-8-9-0-May28
FLEET_URL="<Elastic Agent Fleet URL>"
ENROLLMENT_TOKEN="<Elastic Agent Enrollment Token>"
ELASTIC_ARTIFACT_SERVER="https://artifacts.elastic.co/downloads/beats/elastic-agent" # Replace artifact URL with a pre-release version (BC or snapshot)
ELASTIC_AGENT_VERSION="<Elastic Agent Version>" # e.g: 8.8.0 | 8.8.0-SNAPSHOT

DEV.ALLOW_SSH=false # Set to true to allow SSH connections to the deployed instance
DEV.KEY_NAME="" # When SSH is allowed, you must provide the key name that will be used to ssh into the EC2
>>>>>>> d9dac83 (Accept ElasticArtifactServer as a CloudFormation parameter (#939))
```
