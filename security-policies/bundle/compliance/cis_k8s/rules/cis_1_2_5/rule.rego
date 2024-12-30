package compliance.cis_k8s.rules.cis_1_2_5

import data.compliance.policy.process.ensure_appropriate_arguments as audit
import future.keywords.if

finding := result if {
	audit.apiserver_filter
	result := audit.finding([
		"--kubelet-client-certificate",
		"--kubelet-client-key",
	])
}
