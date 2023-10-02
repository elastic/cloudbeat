package cis_azure.test_data

generate_storage_account_with_property(key, value) = {
	"type": "azure-storage-account",
	"subType": "",
	"resource": {"properties": {key: value}},
}

not_eval_storage_account_empty = {
	"type": "azure-storage-account",
	"subType": "",
	"resource": {"properties": {}},
}

not_eval_non_exist_type = {
	"type": "azure-non-exist",
	"subType": "",
	"resource": {"properties": {}},
}

generate_postgresql_server_with_ssl_enforcement(enabled) = {
	"type": "azure-postgresql-server-db",
	"subType": "",
	"resource": {"properties": {"sslEnforcement": enabled}},
}

generate_mysql_server_with_ssl_enforcement(enabled) = {
	"type": "azure-mysql-server-db",
	"subType": "",
	"resource": {"properties": {"sslEnforcement": enabled}},
}
