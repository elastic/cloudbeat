package compliance.cis_eks.rules.cis_3_2_11

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Verify that the RotateKubeletServerCertificate argument is set to true
default rule_evaluation := false

rule_evaluation if {
	audit.process_contains_key_with_value("--feature-gates", "RotateKubeletServerCertificate=true")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation if {
	audit.not_process_arg_variable("--feature-gates", ["featureGates", "RotateKubeletServerCertificate"])
}

rule_evaluation if {
	audit.not_process_contains_variable("--feature-gates", "RotateKubeletServerCertificate", ["featureGates", "RotateKubeletServerCertificate"])
}

finding := audit.finding(rule_evaluation)
