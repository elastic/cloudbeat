package compliance.cis_k8s.rules.cis_5_1_6

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as test_data.kube_api_pod_input("pod", "name", true)
	test.assert_fail(finding) with input as test_data.kube_api_service_account_input("name", true)
}

test_pass {
	test.assert_pass(finding) with input as test_data.kube_api_pod_input("pod", "name", false)
	test.assert_pass(finding) with input as test_data.kube_api_service_account_input("name", false)
}

test_not_evaluated {
	not finding with input as test_data.not_evaluated_input
	not finding with input as test_data.not_evaluated_kube_api_input
}
