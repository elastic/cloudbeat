package compliance.cis_azure.rules.cis_9_10

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_ftp_disabled),
		data_adapter.resource,
	)
}

is_ftp_disabled {
	is_ftp_enabled != true
} else = false

is_ftp_enabled {
	data_adapter.site_config.ftpsState == "AllAllowed"
} else = false
