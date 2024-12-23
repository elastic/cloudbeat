package compliance.policy.kube_api.minimize_admission_root

import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter
import future.keywords.if

default rule_evaluation := true

# Verify that there is at least one PSP which returns MustRunAsNonRoot or MustRunAs with the range of UIDs not including 0.
rule_evaluation := false if {
	not lib_common.contains_key_with_value(data_adapter.pod.spec.runAsUser, "rule", "MustRunAsNonRoot")
	lib_common.contains_key_with_value(data_adapter.pod.spec.runAsUser, "rule", "MustRunAs")
	range := data_adapter.pod.spec.runAsUser.ranges[_]
	range.min <= 0
}

rule_evaluation := false if {
	container := data_adapter.containers.app_containers[_]
	lib_common.contains_key_with_value(container.securityContext, "runAsUser", 0)
}

finding := result if {
	data_adapter.is_kube_api

	pod := json.filter(data_adapter.pod, [
		"metadata/uid",
		"spec/runAsUser",
	])

	containers := {"containers": {json.filter(c, ["name", "securityContext"]) | c := data_adapter.containers.app_containers[_]}}

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		object.union(pod, containers),
	)
}
