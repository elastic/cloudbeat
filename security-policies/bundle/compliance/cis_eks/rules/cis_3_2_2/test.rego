package compliance.cis_eks.rules.cis_3_2_2

import data.compliance.cis_eks.data_adapter
import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input("")
	eval_fail with input as rule_input("--authorization-mode AlwaysAllow")
	eval_fail with input as rule_input_with_external("--authorization-mode AlwaysAllow", create_process_config("AlwaysAllow"))
	eval_fail with input as rule_input_with_external("", create_process_config("AlwaysAllow"))
}

test_pass if {
	eval_pass with input as rule_input("--authorization-mode Webhook")
	eval_pass with input as rule_input_with_external("--authorization-mode Webhook", create_process_config("AlwaysAllow"))
	eval_pass with input as rule_input_with_external("", create_process_config("Webhook"))
}

test_not_evaluated if {
	not_eval with input as test_data.process_input("some_process", [])
}

rule_input(argument) := test_data.process_input("kubelet", [argument])

rule_input_with_external(argument, external_data) := test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(authz_mode) := {"config": {"authorization": {
	"mode": authz_mode,
	"webhook": {
		"cacheAuthorizedTTL": "0s",
		"cacheUnauthorizedTTL": "0s",
	},
}}}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
