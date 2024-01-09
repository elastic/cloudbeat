package compliance.cis_azure.rules.cis_4_1_3

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# Fail with ServiceManaged
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [{
		"serverKeyType": "ServiceManaged",
		"kind": "azurekeyvault",
		"uri": "any",
	}]})

	# Fail with non azurekeyvault kind
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [{
		"serverKeyType": "AzureKeyVault",
		"kind": "other",
		"uri": "any",
	}]})

	# Fail with empty uri
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [{
		"serverKeyType": "AzureKeyVault",
		"kind": "azurekeyvault",
		"uri": "",
	}]})

	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [{
		"serverKeyType": "AzureKeyVault",
		"kind": "azurekeyvault",
	}]})

	# fail with one good
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [
		{
			"serverKeyType": "AzureKeyVault",
			"kind": "azurekeyvault",
			"uri": "",
		},
		{
			"serverKeyType": "AzureKeyVault",
			"kind": "azurekeyvault",
			"uri": "uri",
		},
	]})

	# fail with without protector
	eval_fail with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {})
}

test_pass if {
	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [{
		"serverKeyType": "AzureKeyVault",
		"kind": "azurekeyvault",
		"uri": "uri",
	}]})

	eval_pass with input as test_data.generate_azure_asset_with_ext("azure-sql-server", {}, {"sqlEncryptionProtectors": [
		{
			"serverKeyType": "AzureKeyVault",
			"kind": "azurekeyvault",
			"uri": "uri",
		},
		{
			"serverKeyType": "AzureKeyVault",
			"kind": "azurekeyvault",
			"uri": "uri",
		},
		{
			"serverKeyType": "AzureKeyVault",
			"kind": "azurekeyvault",
			"uri": "uri",
		},
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
