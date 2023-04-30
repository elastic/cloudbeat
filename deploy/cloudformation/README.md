## Elastic Agent EC2 CloudFormation template

### What it does
This CloudFormation template creates a role for elastic-agent and attaches it to a newly created EC2 instance.
The EC2 instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
The template can be tested with AWS SDK as follows:
Create a `.env` file of the form:
```
STACK_NAME="your unique stack name"
FLEET_URL="<Elastic Agent Fleet URL>"
ENROLLMENT_TOKEN="<Elastic Agent Enrollment Token>"
ELASTIC_AGENT_VERSION="<Elastic Agent Version>" # e.g: 8.8.0 | 8.8.0-SNAPSHOT

DEV.ALLOW_SSH=bool # Set to true in order to modify the template to allow SSH connections
DEV.KEY_NAME="" # When SSH is allowed, your EC2 SSH key is required
DEV.PRE_RELEASE=bool # Set to true in order to replace the artifact URL with a pre-release version (BC or snapshot)
DEV.SHA="" # When running a pre-release version, you have to specify the SHA of the pre-release artifact (on SNAPSHOT versions you can leave empty to take the latest)
```

Run `just deploy-cloudformation`
