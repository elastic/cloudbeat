package compliance.cis_aws.rules.cis_5_1

import data.compliance.policy.aws_ec2.ensure_public_ingress as audit

# Validate that no network acl allow any traffic to remote server admin ports
finding := audit.finding
