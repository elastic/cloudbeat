package compliance.policy.azure.storage_account.ensure_encryption

import data.compliance.policy.azure.data_adapter

is_encryption_enabled {
	data_adapter.properties.encryption.requireInfrastructureEncryption
} else = false
