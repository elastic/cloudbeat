package compliance.policy.azure.storage_account.ensure_public_access

import data.compliance.policy.azure.data_adapter

import future.keywords.if

verify_public_access if {
	data_adapter.properties.publicNetworkAccess == "Disabled"
} else := false

is_public_access_disabled := r if {
	data_adapter.properties.publicNetworkAccess
	r = verify_public_access
}
