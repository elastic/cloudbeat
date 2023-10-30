package compliance.cis_k8s.rules.cis_4_2_3

import data.compliance.policy.process.ensure_arguments_and_config as audit

# Ensure that the --client-ca-file argument is set as appropriate (Automated)
default rule_evaluation = false

rule_evaluation {
	audit.process_contains_key("--client-ca-file")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authentication:x509:clientCAFile: set to a valid path.
rule_evaluation {
	audit.get_from_config(["authentication", "x509", "clientCAFile"])
}

finding = audit.finding(rule_evaluation)
