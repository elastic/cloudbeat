package compliance.cis_k8s.rules.cis_4_2_10

import data.compliance.cis_k8s.data_adapter
import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input("")
	eval_fail with input as rule_input("--tls-cert-file=<path/to/tls-cert>")
	eval_fail with input as rule_input("--tls-private-key-file=<path/to/tls-key>")
}

test_pass if {
	eval_pass with input as rule_input("--tls-private-key-file=<path/to/tls-key> --tls-cert-file=<path/to/tls-cert>")
	eval_pass with input as rule_input_with_external("--tls-private-key-file=<path/to/tls-key> --tls-cert-file=<path/to/tls-cert>", create_process_config("<path/to/tls-key>", "<path/to/tls-cert>"))
	eval_pass with input as rule_input_with_external("", create_process_config("<path/to/tls-key>", "<path/to/tls-cert>"))
}

test_not_evaluated if {
	not_eval with input as test_data.process_input("some_process", [""])
}

rule_input(argument) := test_data.process_input_with_external_data("kubelet", [argument], {})

rule_input_with_external(argument, external_data) := test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(tlsCertFile, tlsPrivateKeyFile) := {"config": {"tlsCertFile": tlsCertFile, "tlsPrivateKeyFile": tlsPrivateKeyFile}}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
