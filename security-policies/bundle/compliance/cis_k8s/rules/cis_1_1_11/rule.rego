package compliance.cis_k8s.rules.cis_1_1_11

import data.compliance.policy.file.ensure_permissions as audit
import future.keywords.if

finding := result if {
	audit.path_filter("/var/lib/etcd/")
	result := audit.finding(audit.file_permission_match(7, 0, 0))
}
