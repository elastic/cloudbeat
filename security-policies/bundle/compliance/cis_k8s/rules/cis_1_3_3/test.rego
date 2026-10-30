package compliance.cis_k8s.rules.cis_1_3_3

import data.compliance.cis_k8s.data_adapter
import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input("")
	eval_fail with input as rule_input("--use-service-account-credentials=false")
	eval_fail with input as rule_input("--use-service-account-credentials=False")
	eval_fail with input as rule_input("--use-service-account-credentials=FALSE")
}

test_pass if {
	eval_pass with input as rule_input("--use-service-account-credentials=true")
	eval_pass with input as rule_input("--use-service-account-credentials=True")
	eval_pass with input as rule_input("--use-service-account-credentials=TRUE")
}

test_not_evaluated if {
	not_eval with input as test_data.process_input("some_process", [])
}

rule_input(argument) := test_data.process_input("kube-controller", [argument])

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
