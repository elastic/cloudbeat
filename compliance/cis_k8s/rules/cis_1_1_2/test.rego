package compliance.cis_k8s.rules.cis_1_1_2

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("kube-apiserver.yaml", "root", "user")
	test.assert_fail(finding) with input as rule_input("kube-apiserver.yaml", "user", "root")
	test.assert_fail(finding) with input as rule_input("kube-apiserver.yaml", "user", "user")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("kube-apiserver.yaml", "root", "root")
}

test_not_evaluated {
	not finding with input as rule_input("file.txt", "root", "root")
}

rule_input(filename, uid, gid) = filesystem_input {
	filemode := "0644"
	filesystem_input = test_data.filesystem_input(filename, filemode, uid, gid)
}
