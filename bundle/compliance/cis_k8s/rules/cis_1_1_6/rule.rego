package compliance.cis_k8s.rules.cis_1_1_6

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.filename_filter("kube-scheduler.yaml")
	result := audit.finding("root", "root")
}
