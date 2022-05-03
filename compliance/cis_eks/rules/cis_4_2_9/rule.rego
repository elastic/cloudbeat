package compliance.cis_eks.rules.cis_4_2_9

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with capabilities assigned (Manual)
# evaluate
default rule_evaluation = true

rule_evaluation = false {
	container := data_adapter.containers[_]
	capabilities := object.get(container.securityContext, "capabilities", [])
	not assert.array_is_empty(capabilities)
}

finding = result {
	# filter
	data_adapter.is_kube_api
	data_adapter.containers

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"containers": {json.filter(c, ["name", "securityContext/capabilities"]) | c := data_adapter.containers[_]},
	}
}
