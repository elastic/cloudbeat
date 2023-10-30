package compliance.cis_k8s.rules.cis_4_2_4

import data.compliance.policy.process.ensure_arguments_and_config as audit

# Verify that the --read-only-port argument is set to 0
default rule_evaluation = false

rule_evaluation {
	audit.process_contains_key_with_value("--read-only-port", "0")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation {
	audit.not_process_arg_comparison("--read-only-port", ["readOnlyPort"], 0)
}

finding = audit.finding(rule_evaluation)
