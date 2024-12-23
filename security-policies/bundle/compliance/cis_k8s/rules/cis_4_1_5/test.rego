package compliance.cis_k8s.rules.cis_4_1_5

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("kubelet.conf", "700")
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("kubelet.conf", "644")
}

test_not_evaluated if {
	not finding with input as rule_input("file.txt", "644")
}

rule_input(filename, filemode) := filesystem_input if {
	user := "root"
	group := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
