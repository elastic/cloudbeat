package compliance.cis_azure.rules.cis_5_1_4

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail when keySource is not Microsoft.Keyvault
	eval_fail with input as test_data.generate_storage_account_with_extensions({}, {"storageAccount": {
		"subscription_id": "subscription_id",
		"id": "storage_account_id",
		"properties": {"encryption": {"keySource": "foo"}},
	}})

	# fail when keyVaultPropertie is null
	eval_fail with input as test_data.generate_storage_account_with_extensions({}, {"storageAccount": {
		"subscription_id": "subscription_id",
		"id": "storage_account_id",
		"properties": {"encryption": {"keySource": "Microsoft.Keyvault"}},
	}})
}

test_pass if {
	# pass when keySource is Microsoft.Keyvault and keyVaultPropertie is not null
	eval_pass with input as test_data.generate_storage_account_with_extensions({}, {"storageAccount": {
		"subscription_id": "subscription_id",
		"id": "storage_account_id",
		"properties": {"encryption": {"keySource": "Microsoft.Keyvault", "keyvaultproperties": {"KeyVaultUri": "key_uri", "keyName": "key_name"}}},
	}})
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
