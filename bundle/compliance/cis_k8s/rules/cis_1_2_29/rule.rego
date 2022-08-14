package compliance.cis_k8s.rules.cis_1_2_29

import data.compliance.policy.process.ensure_arguments_contain_key as audit

finding = result {
	audit.apiserver_filter
	result := audit.finding(audit.contains("--etcd-cafile"))
}
