package compliance.cis_k8s.rules.cis_4_1_6

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("kubelet.conf")
	result := audit.finding("root", "root")
}
