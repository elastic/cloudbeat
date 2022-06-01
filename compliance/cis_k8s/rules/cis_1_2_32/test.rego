package compliance.cis_k8s.rules.cis_1_2_32

import data.kubernetes_common.test_data
import data.lib.test

test_violation {
	test.assert_fail(finding) with input as rule_input("")
	test.assert_fail(finding) with input as rule_input("--tls-cipher-suites=SOME_CIPHER")
	test.assert_fail(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,SOME_CIPHER")
}

test_pass {
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
	test.assert_pass(finding) with input as rule_input("--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
}

test_not_evaluated {
	not finding with input as test_data.process_input("some_process", [])
}

rule_input(argument) = test_data.process_input("kube-apiserver", [argument])
