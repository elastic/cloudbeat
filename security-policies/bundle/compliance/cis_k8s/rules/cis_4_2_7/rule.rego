package compliance.cis_k8s.rules.cis_4_2_7

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --make-iptables-util-chains argument is set to true (Automated)
default rule_evaluation := true

rule_evaluation := false if {
	audit.process_contains_key_with_value("--make-iptables-util-chains", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for makeIPTablesUtilChains is set to true.
rule_evaluation := false if {
	audit.not_process_arg_comparison("--make-iptables-util-chains", ["makeIPTablesUtilChains"], false)
}

finding := audit.finding(rule_evaluation)
