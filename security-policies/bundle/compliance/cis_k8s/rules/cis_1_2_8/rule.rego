package compliance.cis_k8s.rules.cis_1_2_8

import data.compliance.policy.process.ensure_arguments_contain_value as audit

finding := audit.finding(audit.contains("--authorization-mode", "Node"))
