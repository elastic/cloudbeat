package compliance.cis_k8s.rules.cis_1_1_19

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.path_filter("/etc/kubernetes/pki/")
	result := audit.finding("root", "root")
}
