package compliance.cis_k8s.rules.cis_2_5

import data.compliance.policy.process.ensure_arguments_contain_key_value as audit
import future.keywords.if

finding := result if {
	audit.etcd_filter
	result := audit.finding(audit.arg_contains("--peer-client-cert-auth", "true"))
}
