package compliance.cis_azure.rules.cis_4_2_1

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": []})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": [{"properties": {"state": "Disabled"}}]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": [
		{"properties": {"state": "Disabled"}},
		{"properties": {"state": "Enabled"}},
	]})
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": [
		{"properties": {"state": "Disabled"}},
		{"properties": {"state": "Disabled"}},
	]})

	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": [{"properties": {"state": "Enabled"}}]})
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlAdvancedThreatProtectionSettings": [
		{"properties": {"state": "Enabled"}},
		{"properties": {"state": "Enabled"}},
		{"properties": {"state": "Enabled"}},
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
