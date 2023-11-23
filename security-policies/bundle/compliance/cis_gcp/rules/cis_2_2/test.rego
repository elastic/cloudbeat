package compliance.cis_gcp.rules.cis_2_2

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# no sinks
	eval_fail with input as test_data.generate_logging_asset([])

	# sinks with filter
	eval_fail with input as test_data.generate_logging_asset([
		{"resource": {"data": {"name": "log_sink1", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}},
		{"resource": {"data": {"name": "log_sink2", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/folders/123"}},
	])
}

test_pass if {
	# sinks without filter
	eval_pass with input as test_data.generate_logging_asset([{"resource": {"data": {"name": "log_sink1"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}}])

	# project sink with filter and folde sink without filter
	eval_pass with input as test_data.generate_logging_asset([
		{"resource": {"data": {"name": "log_sink1", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}},
		{"resource": {"data": {"name": "log_sink2"}, "parent": "//cloudresourcemanager.googleapis.com/folders/123"}},
	])
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
