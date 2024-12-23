package compliance.policy.process.ensure_appropriate_arguments

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter
import future.keywords.if

process_args := benchmark_data_adapter.process_args

finding(entities) := lib_common.generate_result_without_expected(
	lib_common.calculate_result(rule_evaluation(entities)),
	{"process_args": process_args},
)

# TODO: Change index access to cycle
rule_evaluation(entities) if {
	process_args[entities[0]]
	process_args[entities[1]]
} else := false

apiserver_filter := data_adapter.is_kube_apiserver

etcd_filter := data_adapter.is_etcd
