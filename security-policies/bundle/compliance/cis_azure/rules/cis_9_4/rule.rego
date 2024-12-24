package compliance.cis_azure.rules.cis_9_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_client_cert_enabled),
		data_adapter.resource,
	)
}

is_client_cert_enabled if {
	# To confirm that Client Certificate Mode is set to "Required", both
	# statements have to be true.
	# See: https://github.com/elastic/cloudbeat/issues/1828
	data_adapter.properties.clientCertEnabled == true
	data_adapter.properties.clientCertMode == "Required"
} else := false
