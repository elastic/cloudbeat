package compliance.cis_k8s.rules.cis_2_1

import data.compliance.policy.process.ensure_appropriate_arguments as audit
import future.keywords.if

finding := result if {
	audit.etcd_filter
	result := audit.finding([
		"--cert-file",
		"--key-file",
	])
}
