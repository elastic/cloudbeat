package compliance.cis_azure.rules.cis_8_4

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_key_vault({}, {"vaultSecrets": [test_data.generate_key_vault_extension_key({"enabled": true})]})
	eval_fail with input as test_data.generate_key_vault({"enableRbacAuthorization": false}, {"vaultSecrets": [test_data.generate_key_vault_extension_key({"enabled": true})]})
	eval_fail with input as test_data.generate_key_vault({}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": true}),
	]})
	eval_fail with input as test_data.generate_key_vault({"enableRbacAuthorization": false}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true}),
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
	]})
}

test_pass if {
	eval_pass with input as test_data.generate_key_vault({"enableRbacAuthorization": false}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523298}),
	]})
	eval_pass with input as test_data.generate_key_vault({}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523298}),
	]})
	eval_pass with input as test_data.generate_key_vault({"enableRbacAuthorization": false}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": false}),
	]})
	eval_pass with input as test_data.generate_key_vault({}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": false}),
	]})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_non_exist_type
	not_eval with input as test_data.generate_key_vault({"enableRbacAuthorization": true}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": true, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": false}),
	]})
	not_eval with input as test_data.generate_key_vault({"enableRbacAuthorization": true}, {"vaultSecrets": [
		test_data.generate_key_vault_extension_key({"enabled": false, "exp": 1768523299}),
		test_data.generate_key_vault_extension_key({"enabled": false}),
	]})
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
