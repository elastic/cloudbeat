package compliance.policy.process.data_adapter

import future.keywords.if

is_process if {
	input.type == "process"
}

process_name := name if {
	is_process
	name = input.resource.stat.Name
}

process_args_list := args_list if {
	is_process

	# Gets all the process arguments of the current process
	# Expects format as the following: --<key><delimiter><value> for example: --config=a.json
	# Notice that the first argument is always the process path
	args_list := split(input.resource.command, " --")
}

# Parses a single argument and returns a tuple of the flag and the value
parse_argument(argument) := [flag, value] if {
	# We would like to split the argument by the first delimiter
	# The dilimiter can be either a space or an equal sign
	splitted_argument := regex.split(`\s|\=`, argument)
	flag = concat("", ["--", splitted_argument[0]])

	# We would like to take the entire string after the first delimiter
	value = concat("=", array.slice(splitted_argument, 1, count(splitted_argument) + 1))
}

process_config := config if {
	is_process
	config := {key: value | value = input.resource.external_data[key]}
}

is_kube_apiserver if {
	process_name == "kube-apiserver"
}

is_kube_controller_manager if {
	process_name == "kube-controller"
}

is_kube_scheduler if {
	process_name == "kube-scheduler"
}

is_etcd if {
	process_name == "etcd"
}

is_kubelet if {
	process_name == "kubelet"
}

process_args[flag] := value if {
	[flag, value] = parse_argument(process_args_list[_])
}
