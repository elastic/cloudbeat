package compliance.cis_k8s.rules.cis_4_2_13

import data.compliance.cis_k8s.data_adapter
import data.kubernetes_common.test_data
import data.lib.test

violations {
	test.assert_fail(finding) with input as rule_input("--tls-cipher-suites=SOME_CIPHER")
	test.assert_fail(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,SOME_CIPHER")
	test.assert_fail(finding) with input as rule_input_with_external("", create_process_config(["SOME_CIPHER"]))
	test.assert_fail(finding) with input as rule_input_with_external("", create_process_config(["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,SOME_CIPHER", "SOME_CIPHER"]))
}

test_violation {
	violations with data.benchmark_data_adapter as data_adapter
}

passes {
	test.assert_pass(finding) with input as rule_input("")
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
	test.assert_pass(finding) with input as rule_input_with_external("", create_process_config(["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"]))
	test.assert_pass(finding) with input as rule_input_with_external("", create_process_config(["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"]))
}

test_pass {
	passes with data.benchmark_data_adapter as data_adapter
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", []) with data.benchmark_data_adapter as data_adapter
}

rule_input(argument) = test_data.process_input("kubelet", [argument])

rule_input_with_external(argument, external_data) = test_data.process_input_with_external_data("kubelet", [argument], external_data)

create_process_config(cipherSuites) = {"config": {"TLSCipherSuites": cipherSuites}}
