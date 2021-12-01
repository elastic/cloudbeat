package compliance.cis_k8s.rules.cis_1_4_1

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("scheduler", "")
	test.assert_fail(finding) with input as rule_input("scheduler", "--profiling=true")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("scheduler", "--profiling=false")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.scheduler_input(process_type, [argument])
