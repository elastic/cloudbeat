package compliance.policy.azure.storage_account.ensure_connection

import data.compliance.policy.azure.data_adapter

import future.keywords.every

is_every_private_connections {
	every connection in data_adapter.private_endpoint_connections {
		connection.properties.privateLinkServiceConnectionState.status == "Approved"
	}
} else = false

is_private_connections = r {
	data_adapter.private_endpoint_connections
	r = is_every_private_connections
}
