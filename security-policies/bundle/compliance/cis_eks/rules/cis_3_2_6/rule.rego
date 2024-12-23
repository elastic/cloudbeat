package compliance.cis_eks.rules.cis_3_2_6

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

default rule_evaluation := false

# Ensure that the --protect-kernel-defaults argument is set to true
rule_evaluation if {
	audit.process_contains_key_with_value("--protect-kernel-defaults", "true")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Ensure that the protectKernelDefaults argument is set to true
rule_evaluation if {
	audit.not_process_arg_variable("--protect-kernel-defaults", ["protectKernelDefaults"])
}

finding := audit.finding(rule_evaluation)
