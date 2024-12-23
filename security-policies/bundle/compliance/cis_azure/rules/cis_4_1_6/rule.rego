package compliance.cis_azure.rules.cis_4_1_6

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_retention_long_enough),
		{"Resource": data_adapter.resource},
	)
}

default is_retention_long_enough := false

is_retention_long_enough if {
	data_adapter.resource.extension.sqlBlobAuditPolicy.properties.retentionDays > 90
}

is_retention_long_enough if {
	data_adapter.resource.extension.sqlBlobAuditPolicy.properties.retentionDays == 0 # unlimited retention
}
