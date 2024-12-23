package compliance.cis_azure.rules.cis_6_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_network_watchers_flow_log

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(ensure_retention_days),
		{"Resource": data_adapter.resource},
	)
}

ensure_retention_days if {
	data_adapter.properties.retentionPolicy.enabled
	data_adapter.properties.retentionPolicy.days >= 90
} else := false
