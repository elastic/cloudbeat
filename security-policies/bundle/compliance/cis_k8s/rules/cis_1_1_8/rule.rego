package compliance.cis_k8s.rules.cis_1_1_8

import data.compliance.policy.file.ensure_ownership as audit
import future.keywords.if

finding := result if {
	audit.filename_filter("etcd.yaml")
	result := audit.finding("root", "root")
}
