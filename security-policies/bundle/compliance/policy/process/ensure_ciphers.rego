package compliance.policy.process.ensure_ciphers

import future.keywords.if
import future.keywords.in

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

process_args := benchmark_data_adapter.process_args

is_process_args_includes_non_supported_cipher(supported_ciphers) if {
	ciphers := split(process_args["--tls-cipher-suites"], ",")
	some cipher in ciphers
	not is_supported_cipher(supported_ciphers, cipher)
}

is_process_config_includes_non_supported_cipher(supported_ciphers) if {
	not process_args["--tls-cipher-suites"]
	ciphers := data_adapter.process_config.config.TLSCipherSuites
	cipher := ciphers[_]
	not is_supported_cipher(supported_ciphers, cipher)
}

is_supported_cipher(supported_ciphers, cipher) if {
	cipher in supported_ciphers
}

finding(rule_evaluation) := lib_common.generate_result_without_expected(
	lib_common.calculate_result(rule_evaluation),
	{
		"process_args": process_args,
		"process_config": data_adapter.process_config,
	},
)

apiserver_filter := data_adapter.is_kube_apiserver

kubelet_filter := data_adapter.is_kubelet
