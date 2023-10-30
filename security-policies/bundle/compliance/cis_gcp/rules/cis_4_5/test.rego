package compliance.cis_gcp.rules.cis_4_5

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input({"metadata": {"items": [{"key": "serial-port-enable", "value": "true"}]}})
}

test_pass {
	eval_pass with input as rule_input({})
	eval_pass with input as rule_input({"metadata": {"items": [{"key": "serial-port-enable", "value": "false"}]}})
	eval_pass with input as rule_input({"metadata": {"items": [{"key": "serial-port-enable", "value": 0}]}})
	eval_pass with input as rule_input({"metadata": {"items": [{"key": "dummy-key", "value": "true"}]}})
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

rule_input(info) = test_data.generate_compute_resource("gcp-compute-instance", info)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
