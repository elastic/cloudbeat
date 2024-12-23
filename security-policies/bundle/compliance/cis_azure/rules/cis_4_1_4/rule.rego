package compliance.cis_azure.rules.cis_4_1_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_administrator_configured),
		{"Resource": data_adapter.resource},
	)
}

is_administrator_configured if {
	data_adapter.properties.administrators.administratorType == "ActiveDirectory"
} else := false
