package compliance.cis_eks.rules.cis_3_1_4

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("kubelet-config.json", "root", "user")
	test.assert_fail(finding) with input as rule_input("kubelet-config.json", "user", "root")
	test.assert_fail(finding) with input as rule_input("kubelet-config.json", "user", "user")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("kubelet-config.json", "root", "root")
}

test_not_evaluated {
	not finding with input as rule_input("file.txt", "root", "root")
}

rule_input(filename, user, group) = filesystem_input {
	filemode := "0644"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
