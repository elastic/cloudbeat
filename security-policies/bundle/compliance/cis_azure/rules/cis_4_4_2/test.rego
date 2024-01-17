package compliance.cis_azure.rules.cis_4_4_2

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if no extension
	eval_fail with input as test_data.generate_flexible_mysql_server_with_extension({})

	# fail if no configuration
	eval_fail with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{}]})

	# fail if different configuration
	eval_fail with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "audit_log_enabled",
		"value": "off",
	}]})

	# fail if tls is not v1.2
	eval_fail with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv1.1,tlsv1.0",
	}]})

	# fail if tls is not v1.2
	eval_fail with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv0.9",
	}]})
}

test_pass if {
	# pass if tls version is v.1.2
	eval_pass with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv1.2",
	}]})

	eval_pass with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv1.2,tlsv1.3",
	}]})

	eval_pass with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv1.3,tlsv1.2",
	}]})

	eval_pass with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv1.3,tlsv1.2",
	}]})

	eval_pass with input as test_data.generate_flexible_mysql_server_with_extension({"mysqlConfigurations": [{
		"name": "tls_version",
		"value": "tlsv2.7",
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
