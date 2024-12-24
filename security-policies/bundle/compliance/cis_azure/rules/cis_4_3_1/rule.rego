package compliance.cis_azure.rules.cis_4_3_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_postgresql_single_server_db

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(ssl_enforcement_enabled),
		{"Resource": data_adapter.resource},
	)
}

ssl_enforcement_enabled if {
	lower(data_adapter.properties.sslEnforcement) == "enabled"
} else := false
