package compliance.cis_k8s.rules.cis_2_2

import data.compliance.policy.process.ensure_arguments_contain_key_value as audit

finding = result {
	audit.etcd_filter
	result := audit.finding(audit.contains("--client-cert-auth", "true"))
}
