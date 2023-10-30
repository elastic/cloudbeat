package compliance.cis_gcp.rules.cis_1_8

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "key-management"

subtype := "gcp-cloudresourcemanager-project"

test_violation {
	# fail when same user has both roles (user:a)
	eval_fail with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{},
		{"bindings": [
			{
				"role": "roles/iam.serviceAccountAdmin",
				"members": ["user:a", "user:b"],
			},
			{
				"role": "roles/iam.serviceAccountUser",
				"members": ["user:a", "user:c"],
			},
		]},
	)
}

test_pass {
	# pass when no user has both roles
	eval_pass with input as test_data.generate_gcp_asset(
		type,
		subtype,
		{},
		{"bindings": [
			{
				"role": "roles/iam.serviceAccountAdmin",
				"members": ["user:a", "user:b"],
			},
			{
				"role": "roles/iam.serviceAccountUser",
				"members": ["user:c", "user:d"],
			},
		]},
	)
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.no_policy_resource
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
