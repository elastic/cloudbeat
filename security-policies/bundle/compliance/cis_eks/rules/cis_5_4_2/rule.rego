package compliance.cis_eks.rules.cis_5_4_2

import data.compliance.policy.aws_eks.ensure_private_access as audit

finding = audit.finding(false)
