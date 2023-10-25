package compliance.cis_azure.rules.cis_9_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_client_cert_enabled),
		data_adapter.resource,
	)
}

is_client_cert_enabled {
	# Benchmark and Azure implementation in remediation checks previous implemented value
	# Reading the description and rule metadata, we've decided to check the value of the property clientCertMode
	data_adapter.properties.clientCertMode == "Required"
} else = false
