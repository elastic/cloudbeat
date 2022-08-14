package compliance.policy.process.ensure_appropriate_arguments

import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

finding(entities) = result {
	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation(entities)),
		{"process_args": data_adapter.process_args},
	)
}

# TODO: Change index access to cycle
rule_evaluation(entities) {
	data_adapter.process_args[entities[0]]
	data_adapter.process_args[entities[1]]
} else = false {
	true
}

apiserver_filter := data_adapter.is_kube_apiserver

etcd_filter := data_adapter.is_etcd
