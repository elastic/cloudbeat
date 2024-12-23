package compliance.cis_aws.rules.cis_3_9

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input([])
	eval_fail with input as rule_input(null)
}

test_pass if {
	eval_pass with input as rule_input(["flow_logs"])
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(flow_logs) := test_data.generate_vpc_resource(flow_logs)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
