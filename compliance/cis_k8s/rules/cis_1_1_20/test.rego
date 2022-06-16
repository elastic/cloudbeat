package compliance.cis_k8s.rules.cis_1_1_20

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("/etc/kubernetes/pki/client.crt", "0700")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("/etc/kubernetes/pki/client.crt", "0644")
}

test_not_evaluated {
	not finding with input as rule_input("/etc/kubernetes/pki/client.key", "0644")
}

rule_input(filename, filemode) = filesystem_input {
	user := "root"
	group := "root"
	filesystem_input = test_data.filesystem_input(filename, filemode, user, group)
}
