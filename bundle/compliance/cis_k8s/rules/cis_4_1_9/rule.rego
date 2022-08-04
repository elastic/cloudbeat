package compliance.cis_k8s.rules.cis_4_1_9

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("config.yaml")
	result := audit.finding(6, 4, 4)
}
