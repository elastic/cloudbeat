package compliance.cis_azure.rules.cis_4_3_7

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

# regal ignore:rule-length
test_violation if {
	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [{
		"name": "AllowAllWindowsAzureIps",
		"startIPAddress": "196.203.255.0",
		"endIPAddress": "196.203.255.254",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [{
		"name": "AllowAllWindowsAzureIps",
		"startIPAddress": "0.0.0.0",
		"endIPAddress": "0.0.0.0",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [{
		"name": "randomName",
		"startIPAddress": "0.0.0.0",
		"endIPAddress": "0.0.0.0",
	}]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [
		{
			"name": "randomName",
			"startIPAddress": "0.0.0.0",
			"endIPAddress": "0.0.0.0",
		},
		{
			"name": "randomName",
			"startIPAddress": "196.203.255.0",
			"endIPAddress": "196.203.255.254",
		},
	]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [
		{
			"name": "AllowAllWindowsAzureIps",
			"startIPAddress": "0.0.0.0",
			"endIPAddress": "0.0.0.0",
		},
		{
			"name": "randomName",
			"startIPAddress": "196.203.255.0",
			"endIPAddress": "196.203.255.254",
		},
	]})

	eval_fail with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [
		{
			"name": "randomName",
			"startIPAddress": "196.203.253.0",
			"endIPAddress": "196.203.253.254",
		},
		{
			"name": "AllowAllWindowsAzureIps",
			"startIPAddress": "196.203.255.0",
			"endIPAddress": "196.203.255.254",
		},
	]})
}

test_pass if {
	eval_pass with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [{
		"name": "randomName",
		"startIPAddress": "196.203.255.0",
		"endIPAddress": "196.203.255.254",
	}]})

	eval_pass with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": [
		{
			"name": "randomName",
			"startIPAddress": "196.203.255.0",
			"endIPAddress": "196.203.255.254",
		},
		{
			"name": "randomName2",
			"startIPAddress": "196.203.200.0",
			"endIPAddress": "196.203.200.254",
		},
	]})

	eval_pass with input as test_data.generate_postgresql_server_with_extension({"psqlFirewallRules": []})

	eval_pass with input as test_data.generate_postgresql_server_with_extension({})
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
