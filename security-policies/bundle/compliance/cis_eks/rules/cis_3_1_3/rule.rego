package compliance.cis_eks.rules.cis_3_1_3

import data.compliance.policy.file.ensure_permissions as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("kubelet-config.json")
	result := audit.finding(audit.file_permission_match(6, 4, 4))
}
