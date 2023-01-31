package compliance.policy.kube_api.minimize_assigned_capabilities

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.every

default rule_evaluation = false

rule_evaluation {
	every container in data_adapter.containers.app_containers {
		capabilities_restricted(container)
	}
}

finding := result {
	data_adapter.is_kube_api
	data_adapter.containers

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"containers": {json.filter(c, ["name", "securityContext"]) | c := data_adapter.containers.app_containers[_]}},
	)
}

capabilities_restricted(container) {
	assert.array_is_empty(object.get(container, ["securityContext", "capabilities", "add"], []))
} else {
	object.get(container, ["securityContext", "capabilities", "drop"], []) == ["ALL"]
}
