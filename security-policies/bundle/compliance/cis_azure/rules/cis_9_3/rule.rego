package compliance.cis_azure.rules.cis_9_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_latest_tls_version("1.2")),
		data_adapter.resource,
	)
}

is_latest_tls_version(version) if {
	data_adapter.site_config.minTlsVersion >= version
} else = false
