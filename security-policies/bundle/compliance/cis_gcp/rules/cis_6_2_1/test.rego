package compliance.cis_gcp.rules.cis_6_2_1

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-sql"

subtype := "gcp-sqladmin-instance"

test_violation {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "POSTGRES_15", "settings": {"databaseFlags": [{"name": "log_error_verbosity", "value": "verbose"}]}}}, {})
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "POSTGRES_15", "settings": {"databaseFlags": [{"name": "log_error_verbosity", "value": "terse"}]}}}, {})
}

test_pass {
	# By default log_error_verbosity is set to DEFAULT, therefore a missing databaseFlags is a pass
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "POSTGRES_15"}}, {})
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "POSTGRES_15", "settings": {"databaseFlags": []}}}, {})
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "POSTGRES_15", "settings": {"databaseFlags": [{"name": "log_error_verbosity", "value": "default"}]}}}, {})
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
