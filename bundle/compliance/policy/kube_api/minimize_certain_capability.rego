package compliance.policy.kube_api.minimize_certain_capability

import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter

default rule_evaluation = false

# Verify that there is at least one PSP which returns NET_RAW or ALL.
rule_evaluation {
	# Verify that there is at least one PSP which returns NET_RAW.
	data_adapter.pod.spec.requiredDropCapabilities[_] == "NET_RAW"
}

# or 
rule_evaluation {
	# Verify that there is at least one PSP which returns ALL.
	data_adapter.pod.spec.requiredDropCapabilities[_] == "ALL"
}

finding := result {
	data_adapter.is_kube_api

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		json.filter(data_adapter.pod, [
			"metadata/uid",
			"spec/requiredDropCapabilities",
		]),
	)
}
