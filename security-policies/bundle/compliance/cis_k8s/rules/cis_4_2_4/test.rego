package compliance.cis_k8s.rules.cis_4_2_4

import data.compliance.cis_k8s.data_adapter
import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input("")
	eval_fail with input as rule_input("--read-only-port=10")
	eval_fail with input as rule_input_with_external("--read-only-port=10", create_process_config(0))
	eval_fail with input as rule_input_with_external("", create_process_config(10))
}

test_pass if {
	eval_pass with input as rule_input("--read-only-port=0")
	eval_pass with input as rule_input_with_external("--read-only-port=0", create_process_config(10))
	eval_pass with input as rule_input_with_external("--read-only-port=0", create_process_config(0))
	eval_pass with input as rule_input_with_external("", create_process_config(0))
}

test_not_evaluated if {
	not_eval with input as test_data.process_input("some_process", [])
}

rule_input(argument) := test_data.process_input_with_external_data("kubelet", [argument], {})

rule_input_with_external(argument, external_data) := test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(port) := {"config": {"readOnlyPort": port}}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
