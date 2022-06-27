package compliance.cis_eks.rules.cis_4_2_6

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of root containers (Automated)

# evaluate
default rule_evaluation = true

# Verify that there is at least one PSP which returns MustRunAsNonRoot or MustRunAs with the range of UIDs not including 0.
rule_evaluation = false {
	not common.contains_key_with_value(data_adapter.pod.spec.runAsUser, "rule", "MustRunAsNonRoot")
	common.contains_key_with_value(data_adapter.pod.spec.runAsUser, "rule", "MustRunAs")
	range := data_adapter.pod.spec.runAsUser.ranges[_]
	range.min <= 0
}

rule_evaluation = false {
	container := data_adapter.containers[_]
	common.contains_key_with_value(container.securityContext, "runAsUser", 0)
}

finding = result {
	# filter
	data_adapter.is_kube_api

	# set result
	pod := json.filter(data_adapter.pod, [
		"metadata/uid",
		"spec/runAsUser",
	])

	containers := {"containers": json.filter(c, ["name", "securityContext/runAsUser"]) | c := data_adapter.containers[_]}
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": object.union(pod, containers),
	}
}
