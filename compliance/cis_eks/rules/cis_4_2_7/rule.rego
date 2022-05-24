package compliance.cis_eks.rules.cis_4_2_7

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Minimize the admission of containers with the NET_RAW capability (Automated)

# evaluate
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

finding = result {
	# filter
	data_adapter.is_kube_api

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": json.filter(data_adapter.pod, [
			"metadata/uid",
			"spec/requiredDropCapabilities",
		]),
	}
}
