package compliance.cis_azure.rules.cis_4_4_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_flexible_mysql_server_db

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_tls_version_1_2),
		{"Resource": data_adapter.resource},
	)
}

default is_tls_version_1_2 = false

is_tls_version_1_2 if {
	some i
	data_adapter.resource.extension.mysqlConfigurations[i].name == "tls_version"
	data_adapter.resource.extension.mysqlConfigurations[i].value == "tlsv1.2"
}
