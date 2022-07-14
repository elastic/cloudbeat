package compliance.cis_k8s.rules.cis_5_2_6

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with allowPrivilegeEscalation (Automated)

# evaluate
default rule_evaluation = true

# Verify that there is at least one PSP which does not return true.
rule_evaluation = false {
	container := data_adapter.containers[_]
	common.contains_key_with_value(container.securityContext, "allowPrivilegeEscalation", true)
}

finding = result {
	# filter
	data_adapter.is_kube_api

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"uid": data_adapter.pod.metadata.uid,
			"containers": {json.filter(c, ["name", "securityContext/allowPrivilegeEscalation"]) | c := data_adapter.containers[_]},
		},
	}
}
