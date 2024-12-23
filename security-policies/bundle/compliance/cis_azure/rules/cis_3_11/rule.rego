package compliance.cis_azure.rules.cis_3_11

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account
	_ = data_adapter.resource.extension.blobService

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(soft_delete_is_enabled),
		{"Resource": data_adapter.resource},
	)
}

default soft_delete_is_enabled := false

soft_delete_is_enabled if {
	is_policy_valid(data_adapter.resource.extension.blobService.properties.deleteRetentionPolicy)
	is_policy_valid(data_adapter.resource.extension.blobService.properties.containerDeleteRetentionPolicy)
}

is_policy_valid(policy) if {
	policy.enabled == true
	policy.days > 0
} else := false
