package compliance.policy.azure.data_adapter

resource = input.resource

properties = resource.properties

is_disk {
	input.type == "azure-disk"
}

is_unattached_disk {
	is_disk
	properties.diskState == "Unattached"
}

private_endpoint_connections = properties.privateEndpointConnections

network_acls = properties.networkAcls

is_storage_account {
	input.type == "azure-storage-account"
}

is_storage_account {
	input.type == "azure-classic-storage-account"
}

is_postgresql_server_db {
	input.type == "azure-postgresql-server-db"
}

is_mysql_server_db {
	input.type == "azure-mysql-server-db"
}
