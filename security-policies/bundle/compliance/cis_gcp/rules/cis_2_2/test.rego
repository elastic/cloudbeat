package compliance.cis_gcp.rules.cis_2_2

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	# no sinks
	eval_fail with input as test_data.generate_logging_asset([])

	# sinks with filter
	eval_fail with input as test_data.generate_logging_asset([
		{"resource": {"data": {"name": "log_sink1", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}},
		{"resource": {"data": {"name": "log_sink2", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/folders/123"}},
	])
}

test_pass {
	# sinks without filter
	eval_pass with input as test_data.generate_logging_asset([{"resource": {"data": {"name": "log_sink1"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}}])

	# project sink with filter and folde sink without filter
	eval_pass with input as test_data.generate_logging_asset([
		{"resource": {"data": {"name": "log_sink1", "filter": "logName:syslog AND severity>=ERROR"}, "parent": "//cloudresourcemanager.googleapis.com/projects/123"}},
		{"resource": {"data": {"name": "log_sink2"}, "parent": "//cloudresourcemanager.googleapis.com/folders/123"}},
	])
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
