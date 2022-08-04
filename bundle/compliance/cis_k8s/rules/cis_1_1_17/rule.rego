package compliance.cis_k8s.rules.cis_1_1_17

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("controller-manager.conf")
	result := audit.finding(6, 4, 4)
}
