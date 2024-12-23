package compliance.cis_azure.rules.cis_4_5_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_document_db_database_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_virtual_network_filter_enabled),
		{"Resource": data_adapter.resource},
	)
}

is_virtual_network_filter_enabled if {
	data_adapter.properties.isVirtualNetworkFilterEnabled
} else := false
