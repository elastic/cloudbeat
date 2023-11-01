package compliance.cis_k8s.rules.cis_4_1_10

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.filename_filter("kubelet.conf")
	result := audit.finding("root", "root")
}
