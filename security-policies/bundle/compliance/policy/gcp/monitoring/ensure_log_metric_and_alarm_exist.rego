package compliance.policy.gcp.monitoring.ensure_log_metric_and_alarm_exists

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

finding(filter) := result if {
	# filter
	data_adapter.is_monitoring_asset

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_setup_exists(filter)))
}

is_setup_exists(filter) if {
	some log_metric in input.resource.log_metrics
	log_metric.resource.data.filter == filter
	metric_type := log_metric.resource.data.metricDescriptor.type

	some alert in input.resource.alerts
	alert.resource.data.enabled == true

	some condition in alert.resource.data.conditions
	condition.conditionThreshold.filter == sprintf("metric.type=\"%s\"", [metric_type])
} else := false
