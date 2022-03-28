package compliance.cis_k8s.rules.cis_5_1_6

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("Pod", true)
	test.assert_fail(finding) with input as rule_input("ServiceAccount", true)
}

test_pass {
	test.assert_pass(finding) with input as rule_input("Pod", false)
	test.assert_pass(finding) with input as rule_input("ServiceAccount", false)
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
	not finding with input as test_data.not_evaluated_kube_api_input
}

rule_input(kind, automount_setting) = test_data.kube_api_service_account_input(kind, "name", automount_setting)
