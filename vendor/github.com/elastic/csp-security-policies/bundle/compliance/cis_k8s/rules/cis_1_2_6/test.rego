package compliance.cis_k8s.rules.cis_1_2_6

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("--kubelet-certificate-authority=<ca-string>")
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", [])
}

rule_input(argument) = test_data.process_input("kube-apiserver", [argument])
