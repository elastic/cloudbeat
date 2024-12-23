package compliance.cis_k8s.rules.cis_4_1_2

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("10-kubeadm.conf")
	result := audit.finding("root", "root")
}
