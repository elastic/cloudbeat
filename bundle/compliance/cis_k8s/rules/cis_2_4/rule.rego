package compliance.cis_k8s.rules.cis_2_4

import data.compliance.policy.process.ensure_appropriate_arguments as audit

finding = result {
	audit.etcd_filter
	result := audit.finding([
		"--peer-cert-file",
		"--peer-key-file",
	])
}
