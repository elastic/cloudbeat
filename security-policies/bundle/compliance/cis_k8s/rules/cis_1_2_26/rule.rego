package compliance.cis_k8s.rules.cis_1_2_26

import data.compliance.policy.process.ensure_appropriate_arguments as audit

finding = result {
	audit.apiserver_filter
	result := audit.finding([
		"--etcd-certfile",
		"--etcd-keyfile",
	])
}
