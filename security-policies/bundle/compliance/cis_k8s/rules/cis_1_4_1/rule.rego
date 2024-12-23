package compliance.cis_k8s.rules.cis_1_4_1

import data.compliance.policy.process.ensure_arguments_contain_key_value as audit
import future.keywords.if

finding = result if {
	audit.scheduler_filter
	result := audit.finding(audit.arg_contains("--profiling", "false"))
}
