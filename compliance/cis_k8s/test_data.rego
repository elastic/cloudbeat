package cis_k8s.test_data

# test data generater
filesystem_input(filename, mode, uid, gid) = {
	"type": "file-system",
	"path": sprintf("file/path/%s", [filename]),
	"filename": filename,
	"mode": mode,
	"uid": uid,
	"gid": gid,
}

# Recivies an array of arguments representing the kube-scheduler command
process_input(process_name, arguments) = {
	"type": "process",
	"command": concat(" ", array.concat([process_name], arguments)),
}

kube_api_input(resource) = {
	"type": "kube-api",
	"resource": resource,
}
