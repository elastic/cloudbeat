package compliance.cis_gcp.rules.cis_6_1_3

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_cloud_sql
	data_adapter.is_cloud_my_sql

	result := common.generate_result_without_expected(
		common.calculate_result(is_local_infile_flag_disabled),
		data_adapter.resource,
	)
}

is_local_infile_flag_disabled if {
	flags := data_adapter.resource.data.settings.databaseFlags[_]
	flags.name == "local_infile"
	flags.value == "off"
} else := false
