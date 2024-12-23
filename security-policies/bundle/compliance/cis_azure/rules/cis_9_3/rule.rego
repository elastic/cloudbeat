package compliance.cis_azure.rules.cis_9_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_min_tls_version_1_2),
		data_adapter.resource,
	)
}

default is_min_tls_version_1_2 := false

is_min_tls_version_1_2 if {
	tls_version := data_adapter.resource.extension.siteConfig.properties.MinTLSVersion
	tls_version != null
	common.greater_or_equal(tls_version, 1.2)
}
