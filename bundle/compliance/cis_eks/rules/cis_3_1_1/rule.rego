package compliance.cis_eks.rules.cis_3_1_1

import data.compliance.policy.file.ensure_permissions as audit

finding = result {
	audit.filename_filter("kubeconfig")
	result := audit.finding(6, 4, 4)
}
