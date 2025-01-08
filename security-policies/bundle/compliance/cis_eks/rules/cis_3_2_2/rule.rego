package compliance.cis_eks.rules.cis_3_2_2

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
# If the --authorization-mode argument is present check that it is not set to AlwaysAllow.
default rule_evaluation := false

is_authorization_allow_all if {
	audit.process_arg_not_key_value("--authorization-mode", "--authorization-mode", "AlwaysAllow")
}

rule_evaluation if {
	is_authorization_allow_all
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authorization:mode is not set to AlwaysAllow.
rule_evaluation if {
	not is_authorization_allow_all
	audit.process_filter_variable_multi_comparison(["authorization", "mode"], ["authorization", "mode"], "AlwaysAllow")
}

finding := audit.finding(rule_evaluation)
