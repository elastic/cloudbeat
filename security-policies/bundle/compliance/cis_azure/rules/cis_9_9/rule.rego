package compliance.cis_azure.rules.cis_9_9

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_latest_http),
		data_adapter.resource,
	)
}

is_latest_http {
	data_adapter.site_config.http20Enabled == true
} else = false
