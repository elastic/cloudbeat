package compliance.cis_eks.rules.cis_4_2_7

import data.kubernetes_common.test_data
import data.kubernetes_common.test_data as common_test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {}}]}))
	test.assert_fail(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {"capabilities": {}}}]}))
	test.assert_fail(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {"capabilities": {"add": ["NET_RAW"]}}}]}))
}

test_pass {
	test.assert_pass(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"name": "container_1", "securityContext": {"capabilities": {"drop": ["ALL"]}}}]}))
	test.assert_pass(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {"capabilities": {"drop": ["NET_RAW"]}}}]}))
	test.assert_pass(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {"capabilities": {"drop": ["ALL", "NET_RAW"]}}}]}))
	test.assert_pass(finding) with input as rule_input(common_test_data.pod_security_ctx({"containers": [{"securityContext": {"capabilities": {"drop": ["ALL"]}}}, {"securityContext": {"capabilities": {"drop": ["ALL"]}}}]}))
}

test_not_evaluated {
	not finding with input as {"type": "k8s_object", "resource": {"kind": "Node"}}
}

rule_input(resource) = test_data.kube_api_input(resource)
