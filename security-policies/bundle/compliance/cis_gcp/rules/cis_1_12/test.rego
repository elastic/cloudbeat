package compliance.cis_gcp.rules.cis_1_12

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "key-management"

subtype := "gcp-apikeys-key"

test_violation if {
	# has project level api key
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"name": "projects/some_project"}},
		{},
	)
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"name": "organizations/some_org"}},
		{},
	)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.no_policy_resource
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
