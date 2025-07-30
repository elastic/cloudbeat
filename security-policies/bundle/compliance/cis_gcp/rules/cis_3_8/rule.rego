package compliance.cis_gcp.rules.cis_3_8

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_subnetwork
	not_internal_https_load_balancer

	result := common.generate_evaluation_result(common.calculate_result(is_flow_log_configured))
}

is_flow_log_configured if {
	data_adapter.resource.data.enableFlowLogs == true
	data_adapter.resource.data.logConfig.metadata == "INCLUDE_ALL_METADATA"
	data_adapter.resource.data.logConfig.aggregationInterval == "INTERVAL_5_SEC"
	data_adapter.resource.data.logConfig.flowSampling == 1
	data_adapter.resource.data.logConfig.enable == true
} else := false

not_internal_https_load_balancer if {
	not data_adapter.resource.data.purpose == "INTERNAL_HTTPS_LOAD_BALANCER"
}
