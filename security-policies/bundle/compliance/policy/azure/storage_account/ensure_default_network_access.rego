package compliance.policy.azure.storage_account.ensure_default_network_access

import data.compliance.policy.azure.data_adapter
import future.keywords.if

is_default_network_access_disabled if {
	data_adapter.network_acls.defaultAction == "Deny"
} else := false

is_default_network_access := r if {
	data_adapter.network_acls
	r = is_default_network_access_disabled
}
