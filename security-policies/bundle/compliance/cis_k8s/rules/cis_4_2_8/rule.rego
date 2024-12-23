package compliance.cis_k8s.rules.cis_4_2_8

import data.compliance.policy.process.ensure_arguments_and_config as audit
import future.keywords.if

# Ensure that the --hostname-override argument is not set.
default rule_evaluation := true

# Note This setting is not configurable via the Kubelet config file.
rule_evaluation := false if {
	audit.process_contains_key("--hostname-override")
}

finding := audit.finding(rule_evaluation)
