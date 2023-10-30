package compliance.cis_k8s.rules.cis_1_1_4

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.filename_filter("kube-controller-manager.yaml")
	result := audit.finding("root", "root")
}
