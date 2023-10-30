package compliance.policy.kube_api.minimize_sharing

import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.kube_api.data_adapter

finding(entity) := result {
	data_adapter.is_kube_api

	rule_evaluation := assert.is_false(lib_common.contains_key_with_value(data_adapter.pod.spec, entity, true))

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		json.filter(data_adapter.pod, [
			"metadata/uid",
			concat("", ["spec/", entity]),
		]),
	)
}
