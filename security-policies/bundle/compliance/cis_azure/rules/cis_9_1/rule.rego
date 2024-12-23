package compliance.cis_azure.rules.cis_9_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_app_service_authentication_enabled),
		data_adapter.resource,
	)
}

default is_app_service_authentication_enabled := false

is_app_service_authentication_enabled if {
	data_adapter.resource.extension.authSettings.properties.Enabled == true
}
