package compliance.policy.process.ensure_arguments_contain_key_value

import data.benchmark_data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.common as process_common
import data.compliance.policy.process.data_adapter

process_args := benchmark_data_adapter.process_args

finding(rule_evaluation) := lib_common.generate_result_without_expected(
	lib_common.calculate_result(rule_evaluation),
	{"process_args": process_args},
)

arg_not_contains(entity, value) := assert.is_false(lib_common.contains_key_with_value(process_args, entity, value))

arg_contains(entity, value) := lib_common.contains_key_with_value(process_args, entity, value)

# arg_contains_bool matches a flag whose value is a boolean, using
# lib_common.bool_equal (strconv.ParseBool, case-insensitive).
arg_contains_bool(entity, value) if {
	lib_common.bool_equal(process_args[entity], value)
} else := false

arg_not_contains_bool(entity, value) := assert.is_false(arg_contains_bool(entity, value))

# arg_contains_key_value_bool matches key=value entries (including in a
# comma-separated flag value) with a case-insensitive boolean comparison.
arg_contains_key_value_bool(entity, key, expected) if {
	process_common.arg_value_contains_key_bool(object.get(process_args, entity, ""), key, expected)
} else := false

apiserver_filter := data_adapter.is_kube_apiserver

controller_manager_filter := data_adapter.is_kube_controller_manager

scheduler_filter := data_adapter.is_kube_scheduler

etcd_filter := data_adapter.is_etcd
