package compliance.cis_azure.rules.cis_4_4_1

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test

test_violation {
	# fail if ssl_enforcement is disabled
	eval_fail with input as test_data.generate_mysql_server_with_ssl_enforcement("Disabled")
}

test_pass {
	# pass if ssl_enforcement is enabled
	eval_pass with input as test_data.generate_mysql_server_with_ssl_enforcement("Enabled")
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_non_exist_type
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
