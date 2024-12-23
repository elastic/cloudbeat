package compliance.cis_azure.rules.cis_3_11

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation_container_delete_retention_only if {
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 14),
			generate_container_delete_retention_policy(false, 7),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 14),
			generate_container_delete_retention_policy(true, 0),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 14),
			{"enabled": true},
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 14),
			{},
		),
	)
}

test_violation_delete_retention_only if {
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 0),
			generate_container_delete_retention_policy(true, 7),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(false, 7),
			generate_container_delete_retention_policy(true, 7),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			{"days": 7},
			generate_container_delete_retention_policy(true, 7),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			{"enabled": true},
			generate_container_delete_retention_policy(true, 7),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			{},
			generate_container_delete_retention_policy(true, 7),
		),
	)
}

test_violation_mixed if {
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 5),
			generate_container_delete_retention_policy(true, 0),
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			{},
			{},
		),
	)
	eval_fail with input as test_data.generate_storage_account_with_extensions(
		{},
		{"blobService": {}},
	)
}

test_pass if {
	eval_pass with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 14),
			generate_container_delete_retention_policy(true, 14),
		),
	)
	eval_pass with input as test_data.generate_storage_account_with_extensions(
		{},
		generate_blob_service(
			generate_delete_retention_policy(true, 7),
			generate_container_delete_retention_policy(true, 7),
		),
	)
}

test_not_evaluated if {
	not_eval with input as test_data.generate_storage_account_with_extensions({}, {})
	not_eval with input as test_data.not_eval_storage_account_empty
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

generate_blob_service(delete_retention_policy, container_delete_retention_policy) := {"blobService": {"properties": {
	"deleteRetentionPolicy": delete_retention_policy,
	"containerDeleteRetentionPolicy": container_delete_retention_policy,
}}}

generate_delete_retention_policy(enabled, days) := {
	"enabled": enabled,
	"days": days,
}

generate_container_delete_retention_policy(enabled, days) := {
	"enabled": enabled,
	"days": days,
}
