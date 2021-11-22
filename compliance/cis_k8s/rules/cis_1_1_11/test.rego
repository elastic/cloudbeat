package compliance.cis_k8s.rules.cis_1_1_11

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("var/lib/etcd", "0710")
	test.assert_fail(finding) with input as rule_input("var/lib/etcd/some_file.txt", "0710")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("var/lib/etcd", "0600")
	test.assert_pass(finding) with input as rule_input("var/lib/etcd/some_file.txt", "0600")
}

test_not_evaluated {
	not finding with input as rule_input("file.txt", "0644")
	not finding with input as rule_input("var/lib/etcdd", "0710")
	not finding with input as rule_input("var/lib/etcdd/some_file.txt", "0710")
}

rule_input(filename, filemode) = filesystem_input {
	uid := "root"
	gid := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, uid, gid)
}
