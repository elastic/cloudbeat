package compliance.cis_k8s.rules.cis_5_2_9

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with added capabilities (Automated)

finding = result {
	# filter
	data_adapter.is_kube_api

	# evaluate
	allowedCapabilities := object.get(data_adapter.pod.spec, "allowedCapabilities", [])
	rule_evaluation := assert.array_is_empty(allowedCapabilities)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": json.filter(data_adapter.pod, [
			"metadata/uid",
			"spec/allowedCapabilities",
		]),
	}
}
