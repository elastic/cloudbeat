package compliance.cis_k8s.rules.cis_5_1_3

import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule(["*"], [""], [""])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule([""], ["*"], [""])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule([""], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule(["*"], ["*"], [""])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule([""], ["*"], ["*"])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule(["*"], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule([""], [""], [""]), rule([""], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("ClusterRole", [rule([""], [""], [""]), rule([""], [""], ["create", "*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule(["*"], [""], [""])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], ["*"], [""])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule(["*"], ["*"], [""])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], ["*"], ["*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule(["*"], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], [""], ["create", "*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], [""], [""]), rule([""], [""], ["*"])])
	test.assert_fail(finding) with input as rule_input("Role", [rule([""], [""], [""]), rule([""], [""], ["create", "*"])])
}

test_pass if {
	test.assert_pass(finding) with input as rule_input("ClusterRole", [rule([""], [""], [""])])
	test.assert_pass(finding) with input as rule_input("ClusterRole", [rule(["create"], [""], [""])])
	test.assert_pass(finding) with input as rule_input("ClusterRole", [rule([""], ["create"], [""])])
	test.assert_pass(finding) with input as rule_input("ClusterRole", [rule([""], [""], ["create"])])
	test.assert_pass(finding) with input as rule_input("ClusterRole", [rule([""], [""], ["create"]), rule([""], [""], ["create"])])
	test.assert_pass(finding) with input as rule_input("Role", [rule([""], [""], [""])])
	test.assert_pass(finding) with input as rule_input("Role", [rule(["create"], [""], [""])])
	test.assert_pass(finding) with input as rule_input("Role", [rule([""], ["create"], [""])])
	test.assert_pass(finding) with input as rule_input("Role", [rule([""], [""], ["create"])])
	test.assert_pass(finding) with input as rule_input("Role", [rule([""], [""], ["create", ""])])
	test.assert_pass(finding) with input as rule_input("Role", [rule([""], [""], ["create", ""]), rule([""], [""], ["create", ""])])
}

test_not_evaluated if {
	not finding with input as test_data.not_evaluated_input
	not finding with input as test_data.not_evaluated_kube_api_input
}

rule_input(kind, rules) := test_data.kube_api_role_input(kind, rules)

rule(api_group, resource, verb) := test_data.kube_api_role_rule(api_group, resource, verb)
