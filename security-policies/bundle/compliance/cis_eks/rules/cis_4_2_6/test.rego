package compliance.cis_eks.rules.cis_4_2_6

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input(violating_psp)
	test.assert_fail(finding) with input as rule_input(violating_psp2)
	test.assert_fail(finding) with input as rule_input(violating_psp3)
}

test_pass if {
	test.assert_pass(finding) with input as rule_input(non_violating_psp)
	test.assert_pass(finding) with input as rule_input(non_violating_psp2)
}

test_not_evaluated if {
	not finding with input as {"type": "no-kube-api"}
}

rule_input(resource) := test_data.kube_api_input(resource)

violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"runAsUser": {
		"rule": "MustRunAs",
		"ranges": [{
			"min": 0,
			"max": 65535,
		}],
	}},
}

violating_psp2 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"name": "container_1", "securityContext": {"runAsUser": 0}}]},
}

violating_psp3 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"runAsUser": {
			"rule": "MustRunAs",
			"ranges": [{
				"min": 1,
				"max": 65535,
			}],
		},
		"containers": [{"name": "container_1", "securityContext": {"runAsUser": 0}}],
	},
}

non_violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"runAsUser": {
		"rule": "MustRunAs",
		"ranges": [{
			"min": 1,
			"max": 65535,
		}],
	}},
}

non_violating_psp2 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"runAsUser": {"rule": "MustRunAsNonRoot"}},
}
