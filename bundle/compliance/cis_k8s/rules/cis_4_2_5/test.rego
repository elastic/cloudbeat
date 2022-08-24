package compliance.cis_k8s.rules.cis_4_2_5

import data.compliance.cis_k8s.data_adapter
import data.kubernetes_common.test_data
import data.lib.test

violations {
	test.assert_fail(finding) with input as rule_input("--streaming-connection-idle-timeout=0")
	test.assert_fail(finding) with input as rule_input_with_external("--streaming-connection-idle-timeout=0", create_process_config(0))
	test.assert_fail(finding) with input as rule_input_with_external("--streaming-connection-idle-timeout=0", create_process_config(10))
	test.assert_fail(finding) with input as rule_input_with_external("", create_process_config(0))
}

test_violation {
	violations with data.benchmark_data_adapter as data_adapter
}

passes {
	test.assert_pass(finding) with input as rule_input("")
	test.assert_pass(finding) with input as rule_input("--streaming-connection-idle-timeout=10")
	test.assert_pass(finding) with input as rule_input_with_external("--streaming-connection-idle-timeout=10", create_process_config(0))
	test.assert_pass(finding) with input as rule_input_with_external("--streaming-connection-idle-timeout=10", create_process_config(10))
	test.assert_pass(finding) with input as rule_input_with_external("", create_process_config(10))
}

test_pass {
	passes with data.benchmark_data_adapter as data_adapter
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", []) with data.benchmark_data_adapter as data_adapter
}

rule_input(argument) = test_data.process_input_with_external_data("kubelet", [argument], {})

rule_input_with_external(argument, external_data) = test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(connection_timeout) = {"config": {"streamingConnectionIdleTimeout": connection_timeout}}
