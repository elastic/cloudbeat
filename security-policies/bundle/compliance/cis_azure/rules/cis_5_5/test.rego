package compliance.cis_azure.rules.cis_5_5

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_azure_sku_asset_with_properties("azure-storage-account", {"tier": "Basic"})
	eval_fail with input as test_data.generate_azure_sku_asset_with_properties("azure-storage-account", {})
	eval_fail with input as test_data.generate_azure_sku_asset_with_properties("azure-any-type", {"tier": "Basic"})
	eval_fail with input as test_data.generate_azure_sku_asset_with_properties("azure-any-type", {})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_sku_asset_with_properties("azure-storage-account", {"tier": "Standard"})
	eval_pass with input as test_data.generate_azure_sku_asset_with_properties("azure-any-type", {"tier": "Standard"})
}

test_not_evaluated if {
	not_eval with input as test_data.generate_azure_non_sku_asset("azure-storage-account")
	not_eval with input as test_data.generate_azure_non_sku_asset("azure-any-type")
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
