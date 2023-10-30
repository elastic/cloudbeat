package compliance.cis_aws.rules.cis_3_9

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input([])
	eval_fail with input as rule_input(null)
}

test_pass {
	eval_pass with input as rule_input(["flow_logs"])
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(flow_logs) = test_data.generate_vpc_resource(flow_logs)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
