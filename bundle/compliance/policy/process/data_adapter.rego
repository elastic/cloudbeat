package compliance.policy.process.data_adapter

is_process {
	input.type == "process"
}

process_name = name {
	is_process
	name = input.resource.stat.Name
}

process_args_list = args_list {
	is_process

	# Gets all the process arguments of the current process
	# Expects format as the following: --<key><delimiter><value> for example: --config=a.json
	# Notice that the first argument is always the process path
	args_list := split(input.resource.command, " --")
}

# # This method creates a process args object
# # The object will contain all the process `flags` and their matching values as object key,value accordingly
# process_args(delimiter) = {flag: value | [flag, value] = parse_argument(process_args_list[_], delimiter)}

parse_argument(argument, delimiter) = [flag, value] {
	splitted_argument = split(argument, delimiter)
	flag = concat("", ["--", splitted_argument[0]])

	# We would like to take the entire string after the first delimiter
	value = concat(delimiter, array.slice(splitted_argument, 1, count(splitted_argument) + 1))
}

process_config = config {
	is_process
	config := {key: value | value = input.resource.external_data[key]}
}

is_kube_apiserver {
	process_name == "kube-apiserver"
}

is_kube_controller_manager {
	process_name == "kube-controller"
}

is_kube_scheduler {
	process_name == "kube-scheduler"
}

is_etcd {
	process_name == "etcd"
}

is_kubelet {
	process_name == "kubelet"
}

# TODO: Audits should be able to handle different processes
# process_args = result {
# 	result = data_adapter.process_args(" ")
# }

# process_args = data_adapter.process_args("=")

# TODO: Code something more complex this works only for k8s
process_args = {flag: value | [flag, value] = parse_argument(process_args_list[_], "=")}
