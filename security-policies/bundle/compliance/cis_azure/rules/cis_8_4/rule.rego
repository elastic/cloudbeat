package compliance.cis_azure.rules.cis_8_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.disk.ensure_expiration as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vault
	not data_adapter.properties.enableRbacAuthorization
	data_adapter.resource.extension.vaultSecrets[_].properties.attributes.enabled == true

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.all_enabled_items_have_expiration(data_adapter.resource.extension.vaultSecrets)),
		{"Resource": data_adapter.resource},
	)
}
