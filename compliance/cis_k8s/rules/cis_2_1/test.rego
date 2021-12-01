package compliance.cis_k8s.rules.cis_2_1

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("etcd", "")
	test.assert_fail(finding) with input as rule_input("etcd", "--cert-file=</path/to/ca-file>")
	test.assert_fail(finding) with input as rule_input("etcd", "--key-file=</path/to/key-file>")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("etcd", "--cert-file=</path/to/ca-file> --key-file=</path/to/key-file>")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.etcd_input(process_type, [argument])
