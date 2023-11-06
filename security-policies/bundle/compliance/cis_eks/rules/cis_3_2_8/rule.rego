package compliance.cis_eks.rules.cis_3_2_8

import data.compliance.policy.process.ensure_arguments_and_config as audit

# Ensure that the --hostname-override argument is not set.
default rule_evaluation = true

# Note This setting is not configurable via the Kubelet config file.
rule_evaluation = false {
	audit.process_contains_key("--hostname-override")
}

finding = audit.finding(rule_evaluation)
