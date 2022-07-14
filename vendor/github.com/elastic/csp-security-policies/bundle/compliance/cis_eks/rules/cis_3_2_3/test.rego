package compliance.cis_eks.rules.cis_3_2_3

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("--client-ca-file <path/to/client-ca-file>")
	test.assert_pass(finding) with input as rule_input_with_external("--client-ca-file <path/to/client-ca-file>", create_process_config("<path/to/client-ca-file>"))
	test.assert_pass(finding) with input as rule_input_with_external("", create_process_config("<path/to/client-ca-file>"))
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", [])
}

rule_input(argument) = test_data.process_input("kubelet", [argument])

rule_input_with_external(argument, external_data) = test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(client_CA_path) = {"config": {"authentication": {
	"x509": {"clientCAFile": client_CA_path},
	"anonymous": {"enabled": false},
	"webhook": {
		"cacheTTL": "0s",
		"enabled": true,
	},
}}}
