package compliance.cis_azure.rules.cis_5_1_5

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if storage account id not defined
	eval_fail with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {"storageAccountId": null}}]})

	# fail if logs category is not audit
	eval_fail with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {
		"storageAccountId": "/subscriptions/1",
		"logs": [{
			"category": "AzurePolicyEvaluationDetails",
			"enabled": true,
		}],
	}}]})

	# fail if logs category is audit, but not enabled
	eval_fail with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {
		"storageAccountId": "/subscriptions/1",
		"logs": [{
			"category": "AuditEvent",
			"enabled": false,
		}],
	}}]})
}

test_pass if {
	# pass if the diagnostic setting has:
	# 1. a storage account id
	# 2. logs category is "AuditEvent" or categoryGroup is "audit"/"allLogs"
	# 3. said log category is enabled
	eval_pass with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {
		"storageAccountId": "/subscription/1",
		"logs": [{
			"category": "AuditEvent",
			"enabled": true,
		}],
	}}]})
	eval_pass with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {
		"storageAccountId": "/subscription/1",
		"logs": [{
			"categoryGroup": "audit",
			"enabled": true,
		}],
	}}]})
	eval_pass with input as test_data.generate_key_vault({}, {"vaultDiagnosticSettings": [{"properties": {
		"storageAccountId": "/subscription/1",
		"logs": [{
			"categoryGroup": "allLogs",
			"enabled": true,
		}],
	}}]})
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
