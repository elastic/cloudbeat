package compliance.cis_k8s.rules.cis_1_1_12

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.path_filter("/var/lib/etcd/")
	result := audit.finding("etcd", "etcd")
}
