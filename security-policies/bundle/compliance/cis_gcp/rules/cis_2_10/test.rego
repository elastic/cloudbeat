package compliance.cis_gcp.rules.cis_2_10

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# Alert is not enabled.
	eval_fail with input as rule_input([{"resource": {"data": {"filter": pattern, "metricDescriptor": {"type": "logging.googleapis.com/user/test1"}}}}, {"resource": {"data": {"filter": "not-the-right-pattern", "metricDescriptor": {"type": "logging.googleapis.com/user/test2"}}}}], [{"resource": {"data": {"conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test1\""}}]}}}])

	# Alert is enabled but not attached
	eval_fail with input as rule_input([{"resource": {"data": {"filter": pattern, "metricDescriptor": {"type": "logging.googleapis.com/user/test1"}}}}, {"resource": {"data": {"filter": "not-the-right-pattern", "metricDescriptor": {"type": "logging.googleapis.com/user/test2"}}}}], [{"resource": {"data": {"conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test1\""}}]}}}, {"resource": {"data": {"enabled": true, "conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test3\""}}]}}}])

	# The alert is enabled, but it is not connected to any metric.
	eval_fail with input as rule_input([{"resource": {"data": {"filter": pattern, "metricDescriptor": {"type": "logging.googleapis.com/user/test1"}}}}], [{"resource": {"data": {"enabled": true, "conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test2\""}}]}}}])

	# The log metric filter is not of the right pattern.
	eval_fail with input as rule_input([{"resource": {"data": {"filter": "not-the-right-pattern", "metricDescriptor": {"type": "logging.googleapis.com/user/test1"}}}}], [{"resource": {"data": {"enabled": true, "conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test1\""}}]}}}])
}

test_pass if {
	eval_pass with input as rule_input([{"resource": {"data": {"filter": pattern, "metricDescriptor": {"type": "logging.googleapis.com/user/test1"}}}}, {"resource": {"data": {"filter": "not-the-right-pattern", "metricDescriptor": {"type": "logging.googleapis.com/user/test2"}}}}], [{"resource": {"data": {"enabled": true, "conditions": [{"conditionThreshold": {"filter": "metric.type=\"logging.googleapis.com/user/test1\""}}]}}}])
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
}

rule_input(log_metrics, alerts) := test_data.generate_monitoring_asset(log_metrics, alerts)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
