package compliance.cis_azure.rules.cis_4_3_6

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": [{
		"name": "log_retention_days",
		"value": "3",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": [{
		"name": "log_retention_days",
		"value": "1",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": [{
		"name": "log_checkpoints",
		"value": "on",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": []})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({})
}

test_pass if {
	eval_pass with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": [{
		"name": "log_retention_days",
		"value": "4",
	}]})

	eval_pass with input as test_data.generate_postgresql_server_with_extension({"psqlConfigurations": [{
		"name": "log_retention_days",
		"value": "99",
	}]})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_non_exist_type
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
