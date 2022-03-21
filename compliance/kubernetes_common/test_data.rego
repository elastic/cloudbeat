package kubernetes_common.test_data

# test data generater

# genrates `file-system` type input data
filesystem_input(filename, mode, uid, gid) = {
	"type": "file-system",
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
