package compliance.cis_eks.rules.cis_5_4_5

import data.compliance.policy.aws_elb.ensure_certificates as audit

# Ensure there Kuberenetes endpoint private access is enabled
finding := audit.finding
