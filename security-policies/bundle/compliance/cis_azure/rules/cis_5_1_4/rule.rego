package compliance.cis_azure.rules.cis_5_1_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

# Ensure the storage account containing the container with activity logs is encrypted with Customer Managed Key
finding := result if {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_customer_managed_key_encrypted),
		evidence,
	)
}

is_customer_managed_key_encrypted if {
	data_adapter.resource.extension.storageAccount.properties.encryption.keySource == "Microsoft.Keyvault"
	data_adapter.resource.extension.storageAccount.properties.encryption.keyvaultproperties != null
} else := false

evidence := {
	"storageAccountId": data_adapter.resource.extension.storageAccount.id,
	"SubscriptionId": data_adapter.resource.extension.storageAccount.subscription_id,
	"Encryption": {
		"KeySource": object.get(data_adapter.resource.extension.storageAccount.properties.encryption, "keySource", null),
		"KeyVaultProperties": object.get(data_adapter.resource.extension.storageAccount.properties.encryption, "keyVaultProperties", {}),
	},
}
