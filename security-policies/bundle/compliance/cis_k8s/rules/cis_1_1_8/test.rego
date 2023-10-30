package compliance.cis_k8s.rules.cis_1_1_8

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("etcd.yaml", "root", "user")
	test.assert_fail(finding) with input as rule_input("etcd.yaml", "user", "root")
	test.assert_fail(finding) with input as rule_input("etcd.yaml", "user", "user")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("etcd.yaml", "root", "root")
}

test_not_evaluated {
	not finding with input as rule_input("file.txt", "root", "root")
}

rule_input(filename, user, group) = filesystem_input {
	filemode := "644"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
