package compliance.cis_azure.rules.cis_9_10

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_ftps_disabled),
		data_adapter.resource,
	)
}

is_ftps_disabled if {
	is_ftps_enabled != true
} else = false

is_ftps_enabled if {
	data_adapter.site_config.ftpsState == "AllAllowed"
} else = false
