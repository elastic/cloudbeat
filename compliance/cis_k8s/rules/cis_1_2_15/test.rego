package compliance.cis_k8s.rules.cis_1_2_15

import data.cis_k8s.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("api_server", "")
	test.assert_fail(finding) with input as rule_input("api_server", "--disable-admission-plugins=NamespaceLifecycle")
	test.assert_fail(finding) with input as rule_input("api_server", "--disable-admission-plugins=PodNodeSelector,NamespaceLifecycle")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("api_server", "--disable-admission-plugins=AlwaysDeny")
	test.assert_pass(finding) with input as rule_input("api_server", "--disable-admission-plugins=PodNodeSelector,AlwaysDeny")
}

test_not_evaluated {
	not finding with input as rule_input("some_process", "")
}

rule_input(process_type, argument) = test_data.api_server_input(process_type, [argument])
