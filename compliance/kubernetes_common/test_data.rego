package kubernetes_common.test_data

# input test data generater

# input data that should not get evaluated
not_evaluated_input = {
	"type": "input",
	"resource": {"kind": "some_kind"},
}

# kube-api input data that should not get evaluated
not_evaluated_kube_api_input = {
	"type": "kube-api",
	"resource": {"kind": "some_kind"},
}

# genrates `file` type input data
filesystem_input(filename, mode, uid, gid) = {
	"type": "file",
	"resource": {
		"path": sprintf("file/path/%s", [filename]),
		"filename": filename,
		"mode": mode,
		"uid": uid,
		"gid": gid,
	},
}

# genrates `process` type input data
process_input(process_name, arguments) = process_input_with_external_data(process_name, arguments, {})

# genrates `process` type input data
process_input_with_external_data(process_name, arguments, external_data) = {
	"type": "process",
	"resource": {
		"command": concat(" ", array.concat([process_name], arguments)),
		"stat": {"Name": process_name},
		"external_data": external_data,
	},
}

kube_api_input(resource) = {
	"type": "kube-api",
	"resource": resource,
}

kube_api_role_rule(api_group, resource, verb) = {
	"apiGroups": api_group,
	"resources": resource,
	"verbs": verb,
}

kube_api_role_input(kind, rules) = {
	"type": "kube-api",
	"resource": {
		"kind": kind,
		"metadata": {"name": "role-name"},
		"rules": rules,
	},
}

kube_api_service_account_input(kind, name, automount_setting) = {
	"type": "kube-api",
	"resource": {
		"kind": kind,
		"automountServiceAccountToken": automount_setting,
		"metadata": {"name": name},
		"spec": {
			"serviceAccount": name,
			"automountServiceAccountToken": automount_setting,
		},
	},
}
