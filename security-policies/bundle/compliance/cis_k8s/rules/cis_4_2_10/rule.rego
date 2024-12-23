package compliance.cis_k8s.rules.cis_4_2_10

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate
default rule_evaluation := false

rule_evaluation if {
	audit.process_arg_multi("--tls-cert-file", "--tls-private-key-file")
}

rule_evaluation if {
	audit.process_variable_multi(["tlsCertFile"], ["tlsPrivateKeyFile"])
}

finding := audit.finding(rule_evaluation)
