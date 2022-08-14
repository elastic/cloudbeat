package compliance.cis_k8s.rules.cis_1_3_4

import data.compliance.policy.process.ensure_arguments_contain_key as audit

finding = result {
	audit.controller_manager_filter
	result := audit.finding(audit.contains("--service-account-private-key-file"))
}
