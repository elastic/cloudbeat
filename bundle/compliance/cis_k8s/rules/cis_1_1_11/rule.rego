package compliance.cis_k8s.rules.cis_1_1_11

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.path_filter("/var/lib/etcd/")
	result := audit.finding(7, 0, 0)
}
