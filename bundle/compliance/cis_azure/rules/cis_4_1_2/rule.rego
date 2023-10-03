package compliance.cis_azure.rules.cis_4_1_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_public_access_disabled),
		{"Resource": data_adapter.resource},
	)
}

is_public_access_disabled {
	data_adapter.properties.publicNetworkAccess == "Disabled"
} else = false
