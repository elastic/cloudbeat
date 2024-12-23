package compliance.policy.azure.storage_account.ensure_encryption

import data.compliance.policy.azure.data_adapter
import future.keywords.if

is_encryption_enabled if {
	data_adapter.properties.encryption.requireInfrastructureEncryption
} else := false
