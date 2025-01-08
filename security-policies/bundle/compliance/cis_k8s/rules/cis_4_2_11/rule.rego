package compliance.cis_k8s.rules.cis_4_2_11

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Verify that the --rotate-certificates argument is not present, or is set to true.
default rule_evaluation := true

rule_evaluation := false if {
	audit.process_contains_key_with_value("--rotate-certificates", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation := false if {
	audit.not_process_arg_comparison("--rotate-certificates", ["rotateCertificates"], false)
}

finding := audit.finding(rule_evaluation)
