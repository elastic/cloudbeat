package compliance.cis_k8s.rules.cis_4_2_13

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers

default rule_evaluation = true

process_args := cis_k8s.data_adapter.process_args

supported_ciphers = [
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
]

# Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers
rule_evaluation = false {
	ciphers := split(process_args["--tls-cipher-suites"], ",")
	cipher := ciphers[_]
	not is_supported_cipher(cipher)
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation = false {
	not process_args["--tls-cipher-suites"]
	ciphers := data_adapter.process_config.config.TLSCipherSuites
	cipher := ciphers[_]
	not is_supported_cipher(cipher)
}

is_supported_cipher(cipher) {
	supported_ciphers[_] == cipher
}

finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"process_args": process_args,
			"process_config": data_adapter.process_config,
		},
	}
}
