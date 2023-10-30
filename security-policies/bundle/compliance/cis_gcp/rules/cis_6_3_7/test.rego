package compliance.cis_gcp.rules.cis_6_3_7

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-sql"

subtype := "gcp-sqladmin-instance"

test_violation {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "SQLSERVER_2019_STANDARD", "settings": {"databaseFlags": [{"name": "contained database authentication", "value": "on"}]}}}, {})
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "SQLSERVER_2019_STANDARD"}}, {})
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "SQLSERVER_2019_STANDARD", "settings": {"databaseFlags": []}}}, {})
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"databaseVersion": "SQLSERVER_2019_STANDARD", "settings": {"databaseFlags": [{"name": "contained database authentication", "value": "off"}]}}}, {})
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
