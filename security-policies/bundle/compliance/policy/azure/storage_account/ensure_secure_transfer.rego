package compliance.policy.azure.storage_account.ensure_secure_transfer

import data.compliance.policy.azure.data_adapter

is_secure_transfer_enabled {
	data_adapter.properties.supportsHttpsTrafficOnly
} else = false
