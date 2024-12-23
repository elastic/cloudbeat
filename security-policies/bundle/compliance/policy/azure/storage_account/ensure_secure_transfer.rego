package compliance.policy.azure.storage_account.ensure_secure_transfer

import data.compliance.policy.azure.data_adapter
import future.keywords.if

is_secure_transfer_enabled if {
	data_adapter.properties.supportsHttpsTrafficOnly
} else := false
