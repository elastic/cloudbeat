package compliance.cis_k8s.rules.cis_1_2_32

import data.compliance.policy.process.ensure_ciphers as audit
import future.keywords.if

# Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual)
default rule_evaluation := true

rule_evaluation := false if {
	not audit.process_args["--tls-cipher-suites"]
}

rule_evaluation := false if {
	audit.is_process_args_includes_non_supported_cipher(supported_ciphers)
}

supported_ciphers := [
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

finding := result if {
	audit.apiserver_filter
	result := audit.finding(rule_evaluation)
}
