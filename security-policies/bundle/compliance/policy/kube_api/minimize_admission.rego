package compliance.policy.kube_api.minimize_admission

import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.if

finding(entity) := result if {
	data_adapter.is_kube_api

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation(entity)),
		{
			"uid": data_adapter.pod.metadata.uid,
			"containers": {json.filter(c, ["name", concat("", ["securityContext"])]) | c := data_adapter.containers.app_containers[_]},
			"initContainers": {json.filter(c, ["name", concat("", ["securityContext"])]) | c := data_adapter.containers.init_containers[_]},
			"ephemeralContainers": {json.filter(c, ["name", concat("", ["securityContext"])]) | c := data_adapter.containers.init_containers[_]},
		},
	)
}

rule_evaluation(entity) := false if {
	some container_type # "containers", "init_containers", "ephemeral_containers"
	container := data_adapter.containers[container_type][_]
	lib_common.contains_key_with_value(container.securityContext, entity, true)
} else := true
