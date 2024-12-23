package compliance.cis_eks.rules.cis_3_2_1

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --anonymous-auth argument is set to false (Automated)
default rule_evaluation := false

rule_evaluation if {
	audit.process_contains_key_with_value("--anonymous-auth", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authentication:anonymous: enabled set to false.
rule_evaluation if {
	audit.not_process_arg_comparison("--anonymous-auth", ["authentication", "anonymous", "enabled"], false)
}

finding := audit.finding(rule_evaluation)
