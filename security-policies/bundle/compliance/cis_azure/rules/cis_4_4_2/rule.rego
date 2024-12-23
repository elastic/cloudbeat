package compliance.cis_azure.rules.cis_4_4_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_flexible_mysql_server_db

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(contains_tls_version_higher_than_1_2),
		{"Resource": data_adapter.resource},
	)
}

default contains_tls_version_higher_than_1_2 := false

contains_tls_version_higher_than_1_2 if {
	some i
	data_adapter.resource.extension.mysqlConfigurations[i].name == "tls_version"
	is_list_of_versions_higher(data_adapter.resource.extension.mysqlConfigurations[i].properties.value)
}

is_list_of_versions_higher(version) if {
	versions := split(version, ",")
	some i
	clean_version = trim(versions[i], "tlsv")
	chunks := split(clean_version, ".")
	is_version_higher(to_number(chunks[0]), to_number(chunks[1]))
}

is_version_higher(major, minor) if {
	major == 1
	minor >= 2
}

is_version_higher(major, _) if {
	major > 1
}
