package compliance.cis_k8s.rules.cis_1_1_14

import data.compliance.policy.file.ensure_ownership as audit

finding = result {
	audit.filename_filter("admin.conf")
	result := audit.finding("root", "root")
}
