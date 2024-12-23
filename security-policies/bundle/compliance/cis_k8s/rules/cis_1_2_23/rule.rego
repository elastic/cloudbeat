package compliance.cis_k8s.rules.cis_1_2_23

import data.compliance.policy.process.ensure_arguments_lte as audit

finding := audit.finding("--request-timeout", "60s")
