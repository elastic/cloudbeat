package compliance.policy.azure.storage_account.ensure_connection

import data.compliance.policy.azure.data_adapter
import future.keywords.if

import future.keywords.every

is_every_private_connections if {
	# Azure implemented it differently (like previous version of this file)
	# Simplified and implemented exactly like the PDF audit
	count(data_adapter.private_endpoint_connections) > 0
} else := false

is_private_connections := r if {
	data_adapter.private_endpoint_connections
	r = is_every_private_connections
}
