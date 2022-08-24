package compliance.cis_eks.rules.cis_3_2_7

import data.compliance.policy.process.ensure_arguments_and_config as audit

default rule_evaluation = true

# Ensure that the --make-iptables-util-chains argument is set to true (Automated)
rule_evaluation = false {
	audit.process_contains_key_with_value("--make-iptables-util-chains", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for makeIPTablesUtilChains is set to true.
rule_evaluation = false {
	audit.not_process_arg_comparison("--make-iptables-util-chains", ["makeIPTablesUtilChains"], false)
}

finding = audit.finding(rule_evaluation)
