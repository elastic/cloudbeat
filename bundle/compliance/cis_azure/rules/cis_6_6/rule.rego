package compliance.cis_azure.rules.cis_6_6

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_network_watcher

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(ensure_enabled),
		{"Resource": data_adapter.resource},
	)
}

ensure_enabled {
	data_adapter.properties.provisioningState == "Succeeded"
} else = false
