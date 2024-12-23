package compliance.cis_k8s.rules.cis_1_1_12

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/", "root", "root")
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/", "etcd", "root")
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/", "root", "etcd")
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/some_file.txt", "root", "etcd")
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("var/lib/etcd/", "etcd", "etcd")
	test.assert_pass(finding) with input as rule_input("var/lib/etcd/some_file.txt", "etcd", "etcd")
}

test_not_evaluated if {
	not finding with input as rule_input("file.txt", "root", "root")
	not finding with input as rule_input("var/lib/etcdd", "root", "root")
	not finding with input as rule_input("var/lib/etcdd/some_file.txt", "root", "root")
}

rule_input(filename, user, group) := filesystem_input if {
	filemode := "644"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
