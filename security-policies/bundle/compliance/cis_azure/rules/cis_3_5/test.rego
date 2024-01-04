package compliance.cis_azure.rules.cis_3_5

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_fail_all_present if {
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": true},
			{"category": "StorageWrite", "enabled": true},
			{"category": "StorageDelete", "enabled": false},
		]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": true},
			{"category": "StorageWrite", "enabled": false},
			{"category": "StorageDelete", "enabled": true},
		]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": false},
			{"category": "StorageWrite", "enabled": true},
			{"category": "StorageDelete", "enabled": true},
		]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": false},
			{"category": "StorageWrite", "enabled": false},
			{"category": "StorageDelete", "enabled": true},
		]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": false},
			{"category": "StorageWrite", "enabled": false},
			{"category": "StorageDelete", "enabled": false},
		]}}},
	)
}

test_fail if {
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": true},
			{"category": "StorageWrite", "enabled": true},
		]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [{"category": "StorageRead", "enabled": true}]}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": []}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {}}},
	)

	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{},
	)
}

test_pass if {
	eval_pass with input as test_data.generate_storage_account_with_extensions(
		{},
		{"queueDiagnosticSettings": {"properties": {"logs": [
			{"category": "StorageRead", "enabled": true},
			{"category": "StorageWrite", "enabled": true},
			{"category": "StorageDelete", "enabled": true},
		]}}},
	)
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
