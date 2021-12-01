package compliance.cis_k8s.rules.cis_4_2_2

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("kublet", "")
	test.assert_fail(finding) with input as rule_input("kublet", "--authorization-mode=AlwaysAllow")
	test.assert_fail(finding) with input as rule_input("kublet", "--authorization-mode=RBAC,AlwaysAllow")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("kublet", "--authorization-mode=RBAC")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.kublet_input(process_type, [argument])
