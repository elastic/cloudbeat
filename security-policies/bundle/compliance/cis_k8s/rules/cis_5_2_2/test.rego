package compliance.cis_k8s.rules.cis_5_2_2

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input(violating_psp)
	test.assert_fail(finding) with input as rule_input(violating_psp2)
	test.assert_fail(finding) with input as rule_input(violating_psp3)
	test.assert_fail(finding) with input as rule_input(violating_psp4)
	test.assert_fail(finding) with input as rule_input(violating_psp5)
	test.assert_fail(finding) with input as rule_input(violating_psp6)
}

test_pass if {
	test.assert_pass(finding) with input as rule_input(non_violating_psp)
	test.assert_pass(finding) with input as rule_input(non_violating_psp2)
	test.assert_pass(finding) with input as rule_input(non_violating_psp3)
	test.assert_pass(finding) with input as rule_input(non_violating_psp4)
}

test_not_evaluated if {
	not finding with input as {"type": "no-kube-api"}
}

rule_input(resource) := test_data.kube_api_input(resource)

violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"name": "container_1", "securityContext": {"privileged": true}}]},
}

violating_psp2 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [
		{"name": "container_1", "securityContext": {"privileged": true}},
		{"name": "container_2", "securityContext": {"privileged": false}},
	]},
}

violating_psp3 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [
		{"name": "container_1", "securityContext": {"privileged": true}},
		{"name": "container_2", "securityContext": {}},
	]},
}

violating_psp4 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [
		{"name": "container_1", "securityContext": {"privileged": true}},
		{"name": "container_2", "securityContext": {}},
		{"name": "container_3"},
	]},
}

violating_psp5 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"containers": [
			{"name": "container_1", "securityContext": {"privileged": true}},
			{"name": "container_2", "securityContext": {}},
			{"name": "container_3"},
		],
		"initContainers": [{"name": "init_container_1", "securityContext": {}}],
	},
}

violating_psp6 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"containers": [
			{"name": "container_1"},
			{"name": "container_2", "securityContext": {}},
		],
		"initContainers": [{"name": "init_container_1", "securityContext": {"privileged": true}}],
	},
}

violating_psp7 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"containers": [
			{"name": "container_1"},
			{"name": "container_2", "securityContext": {}},
		],
		"initContainers": [{"name": "init_container_1", "securityContext": {}}],
		"ephemeralContainers": [{"name": "ephemeral_container_1", "securityContext": {"privileged": true}}],
	},
}

non_violating_psp := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"name": "container_1", "securityContext": {"privileged": false}}]},
}

non_violating_psp2 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {"containers": [{"name": "container_1", "securityContext": {}}]},
}

non_violating_psp3 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"containers": [{"name": "container_1", "securityContext": {}}],
		"initContainers": [{"name": "init_container_1", "securityContext": {}}],
	},
}

non_violating_psp4 := {
	"kind": "Pod",
	"metadata": {"uid": "00000aa0-0aa0-00aa-00aa-00aa000a0000"},
	"spec": {
		"containers": [{"name": "container_1", "securityContext": {}}],
		"initContainers": [{"name": "init_container_1", "securityContext": {}}],
		"ephemeralContainers": [{"name": "ephemeral_container_1", "securityContext": {}}],
	},
}
