package compliance.cis_k8s.rules.cis_4_2_12

import data.compliance.policy.process.ensure_arguments_and_config as audit

# Verify that the RotateKubeletServerCertificate argument is set to true
default rule_evaluation = false

rule_evaluation {
	audit.not_process_contains_key_with_value("--feature-gates", "RotateKubeletServerCertificate=false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation {
	audit.not_process_arg_variable("--feature-gates", ["featureGates", "RotateKubeletServerCertificate"])
}

rule_evaluation {
	audit.not_process_contains_variable("--feature-gates", "RotateKubeletServerCertificate", ["featureGates", "RotateKubeletServerCertificate"])
}

rule_evaluation {
	audit.process_contains_key_with_value("--rotate-server-certificates", "true")
}

rule_evaluation {
	audit.get_from_config(["serverTLSBootstrap"])
}

finding = audit.finding(rule_evaluation)
