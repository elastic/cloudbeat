package compliance.cis_gcp.rules.cis_5_1

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input(["allUsers", "test.user@google.com"])
	eval_fail with input as rule_input(["allAuthenticatedUsers"])
	eval_fail with input as rule_input(["allUsers", "allAuthenticatedUsers"])
}

test_pass {
	eval_pass with input as {"subType": "gcp-storage-bucket", "resource": {}}
	eval_pass with input as rule_input([])
	eval_pass with input as rule_input(["test.user@google.com"])
	eval_pass with input as rule_input(["test.user@google.com"])
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

rule_input(members) = test_data.generate_gcs_resource(members, false)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
