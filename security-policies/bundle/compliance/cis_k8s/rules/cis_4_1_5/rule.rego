package compliance.cis_k8s.rules.cis_4_1_5

import data.compliance.policy.file.ensure_permissions as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("kubelet.conf")
	result := audit.finding(audit.file_permission_match(6, 4, 4))
}
