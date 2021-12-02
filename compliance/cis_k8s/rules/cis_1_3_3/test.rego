package compliance.cis_k8s.rules.cis_1_3_3

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("controller_manager", "")
	test.assert_fail(finding) with input as rule_input("controller_manager", "--use-service-account-credentials=false")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("controller_manager", "--use-service-account-credentials=true")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.controller_manager_input(process_type, [argument])
