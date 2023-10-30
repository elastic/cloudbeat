package compliance.cis_k8s.rules.cis_1_1_3

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("kube-controller-manager.yaml")
	result := audit.finding(audit.file_permission_match(6, 4, 4))
}
