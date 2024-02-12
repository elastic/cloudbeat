package compliance.cis_azure.rules.cis_2_1_15

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_security_auto_provisioning_settings([{
		"name": "default",
		"properties": {"autoProvision": "Off"},
	}])

	eval_fail with input as test_data.generate_security_auto_provisioning_settings([{
		"name": "default",
		"properties": {},
	}])

	eval_fail with input as test_data.generate_security_auto_provisioning_settings([{
		"name": "non-default",
		"properties": {"autoProvision": "On"},
	}])

	eval_fail with input as test_data.generate_security_auto_provisioning_settings([
		{
			"name": "default",
			"properties": {"autoProvision": "Off"},
		},
		{
			"name": "non-default",
			"properties": {"autoProvision": "On"},
		},
	])
}

test_pass if {
	eval_pass with input as test_data.generate_security_auto_provisioning_settings([{
		"name": "default",
		"properties": {"autoProvision": "On"},
	}])

	eval_pass with input as test_data.generate_security_auto_provisioning_settings([
		{
			"name": "default",
			"properties": {"autoProvision": "On"},
		},
		{
			"name": "non-default",
			"properties": {"autoProvision": "On"},
		},
	])

	eval_pass with input as test_data.generate_security_auto_provisioning_settings([
		{
			"name": "default",
			"properties": {"autoProvision": "On"},
		},
		{
			"name": "non-default",
			"properties": {"autoProvision": "Off"},
		},
	])
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
