## Elastic Agent EC2 CloudFormation template

### What it does
This CloudFormation template creates a role for elastic-agent and attaches it to a newly created EC2 instance.
The EC2 instance has elastic-agent preinstalled in it using the fleet URL and enrollment token.

### How to test it
The template can be tested with AWS CLI as follows:
```
aws cloudformation create-stack --stack-name elastic-agent-ec2          \
    --template-body file://deploy/cloudformation/elastic-agent-ec2.yml  \
    --capabilities CAPABILITY_IAM   \
    --parameters                    \
    ParameterKey=ElasticAgentVersion,ParameterValue=elastic-agent-8.6.0-linux-x86_64  \
    ParameterKey=FleetUrl,ParameterValue=<Elastic Agent Fleet URL>                    \
    ParameterKey=EnrollmentToken,ParameterValue=<Elastic Agent Enrollment Token>
```
