package compliance.cis_azure.rules.cis_4_2_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_defender_on),
		{"Resource": data_adapter.resource},
	)
}

default is_defender_on := false

is_defender_on if {
	count(data_adapter.resource.extension.sqlAdvancedThreatProtectionSettings) > 0

	every setting in data_adapter.resource.extension.sqlAdvancedThreatProtectionSettings {
		setting.properties.state == "Enabled"
	}
}
