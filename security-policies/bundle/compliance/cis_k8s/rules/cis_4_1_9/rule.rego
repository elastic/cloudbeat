package compliance.cis_k8s.rules.cis_4_1_9

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("config.yaml")
	result := audit.finding(audit.file_permission_match(6, 4, 4))
}
