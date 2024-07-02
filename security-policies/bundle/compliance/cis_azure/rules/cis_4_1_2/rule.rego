package compliance.cis_azure.rules.cis_4_1_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_public_access_disabled),
		{"Resource": data_adapter.resource},
	)
}

default sql_access_config_is_permissive = false

sql_access_config_is_permissive if {
	lower(data_adapter.properties.publicNetworkAccess) == "enabled"
}

sql_access_config_is_permissive if {
	some i
	data_adapter.resource.extension.sqlFirewallRules[i].properties.startIpAddress == "0.0.0.0"
}

default is_public_access_disabled = false

is_public_access_disabled if {
	not sql_access_config_is_permissive
}
