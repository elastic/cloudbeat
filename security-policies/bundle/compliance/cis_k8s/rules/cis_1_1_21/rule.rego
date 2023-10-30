package compliance.cis_k8s.rules.cis_1_1_21

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.path_filter("/etc/kubernetes/pki")
	audit.filename_suffix_filter(".key")
	result := audit.finding(audit.file_permission_match_exact(6, 0, 0))
}
