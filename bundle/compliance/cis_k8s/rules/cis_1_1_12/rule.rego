package compliance.cis_k8s.rules.cis_1_1_12

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.path_filter("/var/lib/etcd/")
	result := audit.finding("etcd", "etcd")
}
