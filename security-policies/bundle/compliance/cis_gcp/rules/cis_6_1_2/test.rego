package compliance.cis_gcp.rules.cis_6_1_2

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation {
	# fail when databaseFlags is not set
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype, {"data": {
			"databaseVersion": "MYSQL_8_0_31",
			"settings": {},
		}},
		{},
	)

	# fail when databaseFlags is set to off
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype, {"data": {
			"databaseVersion": "MYSQL_8_0_31",
			"settings": {"databaseFlags": [{"name": "skip_show_database", "value": "off"}]},
		}},
		{},
	)
}

test_pass {
	# pass when local_infile is off
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype, {"data": {
			"databaseVersion": "MYSQL_8_0_31",
			"settings": {"databaseFlags": [{"name": "skip_show_database", "value": "on"}]},
		}},
		{},
	)
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
