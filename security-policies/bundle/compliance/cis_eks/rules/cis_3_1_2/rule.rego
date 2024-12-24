package compliance.cis_eks.rules.cis_3_1_2

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("kubeconfig")
	result := audit.finding("root", "root")
}
