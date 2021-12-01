package compliance.cis_k8s.rules.cis_2_6

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("etcd", "--peer-auto-tls=true")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("etcd", "")
	test.assert_pass(finding) with input as rule_input("etcd", "-peer-auto-tls=false")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.etcd_input(process_type, [argument])
