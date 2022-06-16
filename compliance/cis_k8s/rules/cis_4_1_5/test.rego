package compliance.cis_k8s.rules.cis_4_1_5

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("kubelet.conf", "0700")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("kubelet.conf", "0644")
}

test_not_evaluated {
	not finding with input as rule_input("file.txt", "0644")
}

rule_input(filename, filemode) = filesystem_input {
	user := "root"
	group := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
