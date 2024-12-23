package compliance.cis_k8s.rules.cis_1_1_19

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.path_filter("/etc/kubernetes/pki/")
	result := audit.finding("root", "root")
}
