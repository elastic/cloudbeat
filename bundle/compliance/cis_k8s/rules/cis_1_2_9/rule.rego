package compliance.cis_k8s.rules.cis_1_2_9

import data.compliance.policy.process.ensure_arguments_contain_value as audit

finding = result {
	result := audit.finding(audit.contains("--authorization-mode", "RBAC"))
}
