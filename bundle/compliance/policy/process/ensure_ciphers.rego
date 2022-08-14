package compliance.policy.process.ensure_ciphers

import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

default rule_evaluation = true

# Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual)
rule_evaluation = false {
	not data_adapter.process_args["--tls-cipher-suites"]
}

rule_evaluation = false {
	ciphers := split(data_adapter.process_args["--tls-cipher-suites"], ",")
	cipher := ciphers[_]
	not is_supported_cipher(cipher)
}

is_supported_cipher(cipher) {
	supported_ciphers[_] == cipher
}

supported_ciphers = [
	"TLS_AES_128_GCM_SHA256",
	"TLS_AES_256_GCM_SHA384",
	"TLS_CHACHA20_POLY1305_SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA",
	"TLS_RSA_WITH_AES_128_CBC_SHA",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_RSA_WITH_AES_256_CBC_SHA",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
]

finding = result {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"process_args": data_adapter.process_args,
			"process_config": data_adapter.process_config,
		},
	)
}
