package compliance.cis_k8s.rules.cis_1_2_11

import data.compliance.policy.process.ensure_arguments_contain_value as audit

finding := audit.finding(audit.not_contains("--enable-admission-plugins", "AlwaysAdmit"))
