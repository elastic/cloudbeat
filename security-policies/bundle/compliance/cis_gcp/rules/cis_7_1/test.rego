package compliance.cis_gcp.rules.cis_7_1

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input(["allUsers", "test.user@google.com"])
	eval_fail with input as rule_input(["allAuthenticatedUsers"])
	eval_fail with input as rule_input(["allUsers", "allAuthenticatedUsers"])
}

test_pass {
	eval_pass with input as {"subType": "gcp-bigquery-dataset", "resource": {}}
	eval_pass with input as rule_input([])
	eval_pass with input as rule_input(["test.user@google.com"])
	eval_pass with input as rule_input(["test.user@google.com", "test.user2@google.com"])
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

rule_input(members) = test_data.generate_bq_resource(null, "gcp-bigquery-dataset", members)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
