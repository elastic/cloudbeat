package compliance.cis_k8s.rules.cis_1_2_21

import data.compliance.policy.process.ensure_arguments_goe as audit

finding := audit.finding("--audit-log-maxbackup", 10)
