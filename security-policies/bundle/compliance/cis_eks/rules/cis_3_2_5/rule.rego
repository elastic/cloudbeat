package compliance.cis_eks.rules.cis_3_2_5

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --streaming-connection-idle-timeout argument is not set to 0
default rule_evaluation := true

rule_evaluation := false if {
	audit.process_contains_key_with_value("--streaming-connection-idle-timeout", "0")
}

rule_evaluation := false if {
	audit.process_contains_key_with_value("--streaming-connection-idle-timeout", "0s")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation := false if {
	audit.not_process_arg_comparison("--streaming-connection-idle-timeout", ["streamingConnectionIdleTimeout"], "0")
}

rule_evaluation := false if {
	audit.not_process_arg_comparison("--streaming-connection-idle-timeout", ["streamingConnectionIdleTimeout"], "0s")
}

finding := audit.finding(rule_evaluation)
