package compliance.cis_azure.rules.cis_9_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_https_only),
		data_adapter.resource,
	)
}

is_https_only if {
	data_adapter.properties.httpsOnly == true
} else := false
