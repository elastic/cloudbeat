package compliance.cis_eks.rules.cis_3_1_2

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("kubeconfig", "root", "user")
	test.assert_fail(finding) with input as rule_input("kubeconfig", "user", "root")
	test.assert_fail(finding) with input as rule_input("kubeconfig", "user", "user")
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("kubeconfig", "root", "root")
}

test_not_evaluated if {
	not finding with input as rule_input("file.txt", "root", "root")
}

rule_input(filename, user, group) := filesystem_input if {
	filemode := "644"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
