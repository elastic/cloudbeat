package compliance.policy.azure.data_adapter

# import data.compliance.lib.common

resource = input.resource

properties = resource.properties

private_endpoint_connections = properties.privateEndpointConnections

network_acls = properties.networkAcls

is_storage_account {
	input.type == "azure-storage-account"
}

is_storage_account {
	input.type == "azure-classic-storage-account"
}
