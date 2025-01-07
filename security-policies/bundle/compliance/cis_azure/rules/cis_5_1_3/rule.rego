package compliance.cis_azure.rules.cis_5_1_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account
	data_adapter.resource.extension.usedForActivityLogs == true

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_blob_access_private),
		{"Resource": data_adapter.resource},
	)
}

default is_blob_access_private := false

is_blob_access_private if {
	data_adapter.resource.properties.allowBlobPublicAccess == false
}
