package compliance.cis_k8s.rules.cis_4_2_9

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture
default rule_evaluation := false

rule_evaluation if {
	audit.process_contains_key_with_value("--event-qps", "0")
}

rule_evaluation if {
	audit.not_process_key_comparison("--event-qps", ["eventRecordQPS"], 0)
}

finding := audit.finding(rule_evaluation)
