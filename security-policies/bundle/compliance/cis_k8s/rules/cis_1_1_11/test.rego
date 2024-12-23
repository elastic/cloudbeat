package compliance.cis_k8s.rules.cis_1_1_11

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("var/lib/etcd", "710")
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/some_file.txt", "710")
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("var/lib/etcd", "600")
	test.assert_pass(finding) with input as rule_input("var/lib/etcd/some_file.txt", "600")
}

test_not_evaluated if {
	not finding with input as rule_input("file.txt", "644")
	not finding with input as rule_input("var/lib/etcdd", "710")
	not finding with input as rule_input("var/lib/etcdd/some_file.txt", "710")
}

rule_input(filename, filemode) := filesystem_input if {
	user := "root"
	group := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
