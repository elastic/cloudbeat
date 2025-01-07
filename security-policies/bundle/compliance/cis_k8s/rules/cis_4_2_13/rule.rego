package compliance.cis_k8s.rules.cis_4_2_13

import data.compliance.policy.process.ensure_ciphers as audit
import future.keywords.if

# Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers
default rule_evaluation := true

# Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers
rule_evaluation := false if {
	audit.is_process_args_includes_non_supported_cipher(supported_ciphers)
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation := false if {
	audit.is_process_config_includes_non_supported_cipher(supported_ciphers)
}

supported_ciphers := [
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
]

finding := result if {
	audit.kubelet_filter
	result := audit.finding(rule_evaluation)
}
