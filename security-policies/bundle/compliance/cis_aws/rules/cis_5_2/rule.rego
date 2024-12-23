package compliance.cis_aws.rules.cis_5_2

import data.compliance.policy.aws_ec2.ensure_security_group_public_ingress_ipv4 as audit

finding := audit.finding
