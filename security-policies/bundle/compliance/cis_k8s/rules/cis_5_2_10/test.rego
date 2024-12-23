package compliance.cis_k8s.rules.cis_5_2_10

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input(violating_psp)
}

test_pass if {
	test.assert_pass(finding) with input as rule_input(non_violating_psp)
	test.assert_pass(finding) with input as rule_input(non_violating_psp2)
	test.assert_pass(finding) with input as rule_input(non_violating_psp3)
}

test_not_evaluated if {
	not finding with input as test_data.not_evaluated_input
	not finding with input as test_data.not_evaluated_kube_api_input
}

rule_input(resource) := test_data.kube_api_input(resource)

violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"securityContext": {"capabilities": {"add": ["NET_RAW"]}}}]},
}

non_violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"securityContext": {}}]},
}

non_violating_psp2 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"securityContext": {"capabilities": {}}}]},
}

non_violating_psp3 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"securityContext": {"capabilities": {"drop": ["ALL"]}}}]},
}
