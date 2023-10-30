package compliance.cis_k8s.rules.cis_1_3_5

import data.compliance.policy.process.ensure_arguments_contain_key as audit

finding = result {
	audit.controller_manager_filter
	result := audit.finding(audit.contains("--root-ca-file"))
}
