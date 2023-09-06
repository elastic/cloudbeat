package cis_azure.test_data

generate_storage_account_encryption(enabled) = {
	"type": "azure-storage-account",
	"subType": "",
	"resource": {"properties": {"encryption": {"requireInfrastructureEncryption": enabled}}},
}

not_eval_storage_account_encryption = {
	"type": "azure-storage-account",
	"subType": "",
	"resource": {"properties": {"encryption": {}}},
}
