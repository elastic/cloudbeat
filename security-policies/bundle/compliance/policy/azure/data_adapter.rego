package compliance.policy.azure.data_adapter

resource = input.resource

properties = resource.properties

is_bastion {
	input.subType == "azure-bastion"
}

is_vault {
	input.subType == "azure-vault"
}

bastions = resource

is_disk {
	input.subType == "azure-disk"
}

is_attached_disk {
	is_disk
	properties.diskState == "Attached"
}

is_unattached_disk {
	is_disk
	properties.diskState == "Unattached"
}

is_vm {
	input.subType = "azure-vm"
}

private_endpoint_connections = properties.privateEndpointConnections

network_acls = properties.networkAcls

activity_log_alerts = resource

is_storage_account {
	input.subType == "azure-storage-account"
}

is_activity_log_alerts {
	input.subType == "azure-activity-log-alert"
}

is_storage_account {
	input.subType == "azure-classic-storage-account"
}

is_postgresql_server_db {
	input.subType == "azure-postgresql-server-db"
}

is_mysql_server_db {
	input.subType == "azure-mysql-server-db"
}

is_website_asset {
	input.subType == "azure-web-site"
}

is_network_watchers_flow_log {
	input.subType == "azure-network-watchers-flow-log"
}

is_network_watcher {
	input.subType == "azure-network-watcher"
}

is_sql_server {
	input.subType == "azure-sql-server"
}

is_document_db_database_account {
	input.subType == "azure-document-db-database-account"
}
