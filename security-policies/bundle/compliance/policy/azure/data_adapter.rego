package compliance.policy.azure.data_adapter

import future.keywords.if

resource := input.resource

properties := resource.properties

identity := resource.identity

is_bastion if {
	input.subType == "azure-bastion"
}

is_role_definition if {
	input.subType == "azure-role-definition"
}

is_custom_role_definition if {
	is_role_definition
	properties.type == "CustomRole"
}

is_vault if {
	input.subType == "azure-vault"
}

role_definitions := resource

bastions := resource

is_disk if {
	input.subType == "azure-disk"
}

is_attached_disk if {
	is_disk
	properties.diskState == "Attached"
}

is_unattached_disk if {
	is_disk
	properties.diskState == "Unattached"
}

is_vm if {
	input.subType = "azure-vm"
}

private_endpoint_connections := properties.privateEndpointConnections

network_acls := properties.networkAcls

site_config := properties.siteConfig

activity_log_alerts := resource

diagnostic_settings := resource

is_storage_account if {
	input.subType == "azure-storage-account"
}

is_security_contacts if {
	input.subType == "azure-security-contacts"
}

is_security_auto_provisioning_settings if {
	input.subType == "azure-security-auto-provisioning-settings"
}

is_activity_log_alerts if {
	input.subType == "azure-activity-log-alert"
}

is_storage_account if {
	input.subType == "azure-classic-storage-account"
}

is_diagnostic_settings if {
	input.subType == "azure-diagnostic-settings"
}

is_postgresql_single_server_db if {
	input.subType == "azure-postgresql-server-db"
}

is_postgresql_flexible_server_db if {
	input.subType == "azure-flexible-postgresql-server-db"
}

is_postgresql_server_db if {
	is_postgresql_single_server_db
}

is_postgresql_server_db if {
	is_postgresql_flexible_server_db
}

is_flexible_mysql_server_db if {
	input.subType == "azure-flexible-mysql-server-db"
}

is_mysql_server_db if {
	input.subType == "azure-mysql-server-db"
}

is_website_asset if {
	input.subType == "azure-web-site"
}

is_network_watchers_flow_log if {
	input.subType == "azure-network-watchers-flow-log"
}

is_network_watcher if {
	input.subType == "azure-network-watcher"
}

is_sql_server if {
	input.subType == "azure-sql-server"
}

is_document_db_database_account if {
	input.subType == "azure-document-db-database-account"
}

insights_components := resource

is_insights_component if {
	input.subType == "azure-insights-component"
}
