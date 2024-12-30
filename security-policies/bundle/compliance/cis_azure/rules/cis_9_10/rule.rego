package compliance.cis_azure.rules.cis_9_10

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_website_asset

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_ftp_state_disabled_or_ftps),
		data_adapter.resource,
	)
}

default is_ftp_state_disabled_or_ftps := false

is_ftp_state_disabled_or_ftps if {
	data_adapter.resource.extension.siteConfig.properties.FtpsState == "Disabled"
}

is_ftp_state_disabled_or_ftps if {
	data_adapter.resource.extension.siteConfig.properties.FtpsState == "FtpsOnly"
}
