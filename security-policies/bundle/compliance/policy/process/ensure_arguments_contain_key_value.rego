package compliance.policy.process.ensure_arguments_contain_key_value

import data.benchmark_data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

process_args := benchmark_data_adapter.process_args

finding(rule_evaluation) := lib_common.generate_result_without_expected(
	lib_common.calculate_result(rule_evaluation),
	{"process_args": process_args},
)

arg_not_contains(entity, value) := assert.is_false(lib_common.contains_key_with_value(process_args, entity, value))

arg_contains(entity, value) := lib_common.contains_key_with_value(process_args, entity, value)

apiserver_filter := data_adapter.is_kube_apiserver

controller_manager_filter := data_adapter.is_kube_controller_manager

scheduler_filter := data_adapter.is_kube_scheduler

etcd_filter := data_adapter.is_etcd
