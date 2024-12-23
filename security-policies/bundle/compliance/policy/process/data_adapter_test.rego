package compliance.policy.process.data_adapter

import data.kubernetes_common.test_data
import future.keywords.if

supported_delimiters := [" ", "="]

test_is_process if {
	is_process with input as test_data.process_input("kube-api", [])
}

test_process_name if {
	process_name := "kube-api"
	result := process_name with input as test_data.process_input("kube-api", [])
	result == process_name
}

test_process_args_list if {
	expected_result := ["kube-api", "cloud-provider aws", "config /etc/kubernetes/kubelet/kubelet-config.json"]
	result := process_args_list with input as process_input([], supported_delimiters[_])
	result == expected_result
}

test_process_args_list_when_value_contain_delimiters if {
	some delimiter_index
	delimiter := supported_delimiters[delimiter_index]
	arg_value := replace("--arg%0value%0and%0delimiters", "%0", delimiter)

	## Remove the -- from the begining
	expected_arg_value := trim_prefix(arg_value, "--")
	expected_result := ["kube-api", "cloud-provider aws", "config /etc/kubernetes/kubelet/kubelet-config.json", expected_arg_value]

	result := process_args_list with input as process_input([arg_value], delimiter)
	result == expected_result
}

process_input(extra_elements, delimiter) := test_data.process_input("kube-api", process_cmdLine_input(delimiter, extra_elements))

process_cmdLine_input(delimiter, extra_elements) := result if {
	cmd_line_with_placeholders := ["--cloud-provider%0aws", "--config%0/etc/kubernetes/kubelet/kubelet-config.json"]
	cmd_line_with_extra_elements := array.concat(cmd_line_with_placeholders, extra_elements)
	result = [res | res = replace(cmd_line_with_extra_elements[_], "%0", delimiter)]
}
