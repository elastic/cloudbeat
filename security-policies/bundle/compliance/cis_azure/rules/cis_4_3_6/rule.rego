package compliance.cis_azure.rules.cis_4_3_6

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_postgresql_single_server_db

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(log_retention_long_enough),
		{"Resource": data_adapter.resource},
	)
}

default log_retention_long_enough := false

log_retention_long_enough if {
	some i
	data_adapter.resource.extension.psqlConfigurations[i].name == "log_retention_days"
	to_number(data_adapter.resource.extension.psqlConfigurations[i].properties.value) > 3
}
