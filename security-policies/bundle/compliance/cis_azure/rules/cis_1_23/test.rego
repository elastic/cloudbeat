package compliance.cis_azure.rules.cis_1_23

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": [
			"/",
			"/subscriptions/11111111-2f22-3333-44e4-555f555g5555",
		],
		"permissions": [{"actions": ["*"]}],
		"type": "CustomRole",
	})
	eval_fail with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": ["/subscriptions/11111111-2f22-3333-44e4-555f555g5555"],
		"permissions": [{"actions": ["*"]}],
		"type": "CustomRole",
	})
	eval_fail with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": ["/"],
		"permissions": [{"actions": ["*"]}],
		"type": "CustomRole",
	})
	eval_fail with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": [
			"RandomScope",
			"/",
			"/subscriptions/11111111-2f22-3333-44e4-555f555g5555",
		],
		"permissions": [{"actions": ["RandomAction", "*"]}],
		"type": "CustomRole",
	})
}

test_pass {
	eval_pass with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": [
			"/",
			"/subscriptions/11111111-2f22-3333-44e4-555f555g5555",
		],
		"permissions": [{"actions": ["RandomAction"]}],
		"type": "CustomRole",
	})
	eval_pass with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": [
			"/",
			"/subscriptions/11111111-2f22-3333-44e4-555f555g5555/RandomScope",
		],
		"permissions": [{"actions": ["RandomAction"]}],
		"type": "CustomRole",
	})
	eval_pass with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": ["RandomScope"],
		"permissions": [{"actions": ["*"]}],
		"type": "CustomRole",
	})
}

test_not_evaluated {
	not_eval with input as test_data.generate_azure_asset("azure-role-definition", {
		"assignableScopes": [
			"/",
			"/subscriptions/11111111-2f22-3333-44e4-555f555g5555",
		],
		"permissions": [{"actions": ["*"]}],
		"type": "BuiltInRole",
	})
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
