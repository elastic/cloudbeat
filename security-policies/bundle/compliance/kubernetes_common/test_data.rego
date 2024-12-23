package kubernetes_common.test_data

# input test data generater

# input data that should not get evaluated
not_evaluated_input := {
	"type": "input",
	"resource": {"kind": "some_kind"},
}

# kube-api input data that should not get evaluated
not_evaluated_kube_api_input := {
	"type": "k8s_object",
	"resource": {"kind": "some_kind"},
}

# genrates `file` type input data
filesystem_input(filename, mode, user, group) := {
	"type": "file",
	"resource": {
		"path": sprintf("file/path/%s", [filename]),
		"name": filename,
		"mode": mode,
		"owner": user,
		"group": group,
	},
}

# genrates `process` type input data
process_input(process_name, arguments) := process_input_with_external_data(process_name, arguments, {})

# genrates `process` type input data
process_input_with_external_data(process_name, arguments, external_data) := {
	"type": "process",
	"resource": {
		"command": concat(" ", array.concat([process_name], arguments)),
		"stat": {"Name": process_name},
		"external_data": external_data,
	},
}

kube_api_input(resource) := {
	"type": "k8s_object",
	"resource": resource,
}

kube_api_role_rule(api_group, resource, verb) := {
	"apiGroups": api_group,
	"resources": resource,
	"verbs": verb,
}

kube_api_role_input(kind, rules) := {
	"type": "k8s_object",
	"resource": {
		"kind": kind,
		"metadata": {"name": "role-name"},
		"rules": rules,
	},
}

kube_api_pod_input(pod_name, service_account, automount_setting) := {
	"type": "k8s_object",
	"resource": {
		"kind": "Pod",
		"metadata": {"name": pod_name},
		"spec": {
			"serviceAccount": service_account,
			"serviceAccountName": service_account,
			"automountServiceAccountToken": automount_setting,
		},
	},
}

kube_api_service_account_input(name, automount_setting) := {
	"type": "k8s_object",
	"resource": {
		"kind": "ServiceAccount",
		"metadata": {"name": name},
		"automountServiceAccountToken": automount_setting,
	},
}

pod_security_ctx(entry) := {
	"kind": "Pod",
	"metadata": {"name": "pod-name"},
	"spec": entry,
}
