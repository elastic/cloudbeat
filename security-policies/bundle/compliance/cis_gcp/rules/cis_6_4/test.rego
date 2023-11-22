package compliance.cis_gcp.rules.cis_6_4

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation if {
	# fail when requireSsl doesn't exists
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"databaseVersion": "MYSQL_5_6",
			"settings": {},
		}},
		{},
	)
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"databaseVersion": "MYSQL_5_6",
			"settings": {"ipConfiguration": {}},
		}},
		{},
	)

	# fail when requireSsl is set to false
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"databaseVersion": "MYSQL_5_6",
			"settings": {"ipConfiguration": {"requireSsl": false}},
		}},
		{},
	)
}

test_pass if {
	# pass when requireSsl is set to true
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"databaseVersion": "MYSQL_5_6",
			"settings": {"ipConfiguration": {"requireSsl": true}},
		}},
		{},
	)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"databaseVersion": "SQLSERVER_2019_STANDARD"}},
		{},
	)
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
