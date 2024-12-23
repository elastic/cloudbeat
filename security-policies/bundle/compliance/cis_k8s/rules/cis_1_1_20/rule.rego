package compliance.cis_k8s.rules.cis_1_1_20

import data.compliance.policy.file.ensure_permissions as audit
import future.keywords.if

finding := result if {
	audit.path_filter("/etc/kubernetes/pki")
	audit.filename_suffix_filter(".crt")
	result := audit.finding(audit.file_permission_match(6, 4, 4))
}
