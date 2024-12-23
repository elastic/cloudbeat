package compliance.cis_azure.rules.cis_4_1_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_transaparent_data_encryption_enabled_for_all_dbs),
		{"Resource": data_adapter.resource},
	)
}

default is_transaparent_data_encryption_enabled_for_all_dbs := false

is_transaparent_data_encryption_enabled_for_all_dbs if {
	count(data_adapter.resource.extension.sqlTransparentDataEncryptions) > 0

	every p in data_adapter.resource.extension.sqlTransparentDataEncryptions {
		is_transparent_data_encryption_enabled(p.properties.databaseName, p.properties.state)
	}
}

is_transparent_data_encryption_enabled("master", _) := true

is_transparent_data_encryption_enabled(dbName, state) if {
	dbName != "master"
	state == "Enabled"
}
