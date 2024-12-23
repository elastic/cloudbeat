package compliance.cis_azure.rules.cis_4_1_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_encryption_protector_key_vault),
		{"Resource": data_adapter.resource},
	)
}

default is_encryption_protector_key_vault := false

is_encryption_protector_key_vault if {
	count(data_adapter.resource.extension.sqlEncryptionProtectors) > 0

	every p in data_adapter.resource.extension.sqlEncryptionProtectors {
		p.properties.serverKeyType == "AzureKeyVault"
		p.properties.kind == "azurekeyvault"
		count(trim_space(p.properties.uri)) > 0
	}
}
