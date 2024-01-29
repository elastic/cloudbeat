package compliance.cis_azure.rules.cis_6_6

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# No network watchers
	eval_fail with input as test_data.generate_azure_asset_resource("azure-network-watcher", {})
	eval_fail with input as test_data.generate_azure_asset_resource("azure-network-watcher", {"networkWatchers": []})

	# No succeeded network watcher
	eval_fail with input as test_data.generate_azure_asset_resource("azure-network-watcher", {"networkWatchers": [{"properties": {"provisioningState": "NotSucceeded"}}]})

	eval_fail with input as test_data.generate_azure_asset_resource("azure-network-watcher", {"networkWatchers": [
		{"properties": {"provisioningState": "NotSucceeded"}},
		{"properties": {"provisioningState": "NotSucceeded"}},
	]})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset_resource("azure-network-watcher", {"networkWatchers": [{"properties": {"provisioningState": "Succeeded"}}]})

	eval_pass with input as test_data.generate_azure_asset_resource("azure-network-watcher", {"networkWatchers": [
		{"properties": {"provisioningState": "Succeeded"}},
		{"properties": {"provisioningState": "NotSucceeded"}},
	]})
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
