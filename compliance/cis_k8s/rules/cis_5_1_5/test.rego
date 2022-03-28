package compliance.cis_k8s.rules.cis_5_1_5

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("Pod", "default", false)
	test.assert_fail(finding) with input as rule_input("ServiceAccount", "default", true)
}

test_pass {
	test.assert_pass(finding) with input as rule_input("Pod", "myServiceAccountName", false)
	test.assert_pass(finding) with input as rule_input("Pod", "myServiceAccountName", true)
	test.assert_pass(finding) with input as rule_input("ServiceAccount", "default", false)
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
	not finding with input as test_data.not_evaluated_kube_api_input
}

rule_input(kind, name, automount_setting) = test_data.kube_api_service_account_input(kind, name, automount_setting)
