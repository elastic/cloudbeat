package compliance.cis_k8s.rules.cis_1_2_27

import data.compliance.policy.process.ensure_appropriate_arguments as audit

finding = result {
	audit.apiserver_filter
	result := audit.finding([
		"--tls-cert-file",
		"--tls-private-key-file",
	])
}
