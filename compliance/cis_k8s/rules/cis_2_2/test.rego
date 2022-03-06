package compliance.cis_k8s.rules.cis_2_2

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("")
	test.assert_fail(finding) with input as rule_input("--client-cert-auth=false")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("--client-cert-auth=true")
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", [])
}

rule_input(argument) = test_data.process_input("etcd", [argument])
