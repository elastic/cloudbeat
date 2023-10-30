package compliance.policy.kube_api.minimize_added_capabilities

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter

finding := result {
	data_adapter.is_kube_api

	allowedCapabilities := object.get(data_adapter.pod.spec, "allowedCapabilities", [])
	rule_evaluation := assert.array_is_empty(allowedCapabilities)

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"filemode": json.filter(data_adapter.pod, [
			"metadata/uid",
			"spec/allowedCapabilities",
		])},
	)
}
