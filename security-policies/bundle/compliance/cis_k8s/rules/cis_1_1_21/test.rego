package compliance.cis_k8s.rules.cis_1_1_21

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("/etc/kubernetes/pki/client.key", "700")
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("/etc/kubernetes/pki/client.key", "600")
}

test_not_evaluated if {
	not finding with input as rule_input("/etc/kubernetes/pki/client.crt", "600")
}

rule_input(filename, filemode) := filesystem_input if {
	user := "root"
	group := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
