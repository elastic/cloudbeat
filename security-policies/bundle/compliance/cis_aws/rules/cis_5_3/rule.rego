package compliance.cis_aws.rules.cis_5_3

import data.compliance.policy.aws_ec2.ensure_security_group_public_ingress_ipv6 as audit

finding := audit.finding
