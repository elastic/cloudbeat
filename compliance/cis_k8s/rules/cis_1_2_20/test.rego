package compliance.cis_k8s.rules.cis_1_2_20

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("api_server", "--secure-port=0")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("api_server", "--secure-port=1")
	test.assert_pass(finding) with input as rule_input("api_server", "--secure-port=65535")
	test.assert_pass(finding) with input as rule_input("api_server", "")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.api_server_input(process_type, [argument])
