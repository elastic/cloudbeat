package compliance.cis_k8s.rules.cis_4_1_5

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("kubelet.conf")
	result := audit.finding(6, 4, 4)
}
