package compliance.cis_gcp.rules.cis_1_6

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "key-management"

subtype := "gcp-cloudresourcemanager-project"

test_violation if {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [{
		"role": "roles/iam.serviceAccountUser",
		"members": ["user:c"],
	}]})
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [{
		"role": "roles/iam.serviceAccountTokenCreator",
		"members": ["user:c"],
	}]})
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {}, {"bindings": [{
		"role": "roles/some_other_role",
		"members": ["user:c"],
	}]})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource

	# don't evaluate when resource.policy is missing
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
