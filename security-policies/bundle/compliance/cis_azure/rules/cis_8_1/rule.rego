package compliance.cis_azure.rules.cis_8_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_vault
	data_adapter.properties.enableRbacAuthorization

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(are_enabled_keys_expirable),
		{"Resource": data_adapter.resource},
	)
}

are_enabled_keys_expirable if {
	enabledKeys = [key | key := data_adapter.resource.extension.vaultKeys[_]; key.properties.attributes.enabled == true]

	every key in enabledKeys {
		key.properties.attributes.exp > 0
	}
} else = false
