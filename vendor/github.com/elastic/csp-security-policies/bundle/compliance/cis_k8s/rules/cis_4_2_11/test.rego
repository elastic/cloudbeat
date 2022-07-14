package compliance.cis_k8s.rules.cis_4_2_11

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("--rotate-certificates=false")
	test.assert_fail(finding) with input as rule_input_with_external("", create_process_config(false))
	test.assert_fail(finding) with input as rule_input_with_external("--rotate-certificates=false", create_process_config(true))
}

test_pass {
	test.assert_pass(finding) with input as rule_input("")
	test.assert_pass(finding) with input as rule_input("--rotate-certificates=true")
	test.assert_pass(finding) with input as rule_input_with_external("--rotate-certificates=true", create_process_config(false))
	test.assert_pass(finding) with input as rule_input_with_external("--rotate-certificates=true", create_process_config(true))
	test.assert_pass(finding) with input as rule_input_with_external("", create_process_config(true))
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", [])
}

rule_input(argument) = test_data.process_input("kubelet", [argument])

rule_input_with_external(argument, external_data) = test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(rotateCertificates) = {"config": {"rotateCertificates": rotateCertificates}}
