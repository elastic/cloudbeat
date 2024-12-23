package compliance.cis_k8s.rules.cis_1_2_25

import data.compliance.policy.process.ensure_arguments_contain_key as audit
import future.keywords.if

finding := result if {
	audit.apiserver_filter
	result := audit.finding(audit.arg_contains("--service-account-key-file"))
}
