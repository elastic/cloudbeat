package compliance.policy.kube_api.minimize_certain_capability

import future.keywords.if
import future.keywords.in

import data.compliance.lib.common
import data.compliance.policy.kube_api.data_adapter

default rule_evaluation := false

rule_evaluation if {
	container := data_adapter.containers.app_containers[_]
	"NET_RAW" in container.securityContext.capabilities.drop
}

# or
rule_evaluation if {
	container := data_adapter.containers.app_containers[_]
	"ALL" in container.securityContext.capabilities.drop
}

finding := result if {
	data_adapter.is_kube_pod

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"containers": {json.filter(c, ["name", "securityContext"]) | c := data_adapter.containers.app_containers[_]}},
	)
}
