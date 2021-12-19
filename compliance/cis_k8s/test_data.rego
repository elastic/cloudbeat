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

## Recivies an array of arguments representing the kube-scheduler command
#process_input(process_name, arguments) = {
#	"type": "process",
#	"command": concat(" ", array.concat(["kube-scheduler"], arguments)),
#}
#
## Recivies an array of arguments representing the kube-controller-manager command
#controller_manager_input( arguments) = {
#	"type": process,
#	"command": concat(" ", array.concat(["kube-controller-manager"], arguments)),
#}
#
## Recivies an array of arguments representing the API Server command
#api_server_input(arguments) = {
#	"type": "process",
#	"command": concat(" ", array.concat(["kube-apiserver"], arguments)),
#}
#
## Recivies an array of arguments representing the kube-controller-manager command
#etcd_input(arguments) = {
#	"type": process,
#	"command": concat(" ", array.concat(["etcd"], arguments)),
#}
#
## Recivies an array of arguments representing the kubelet command
#kublet_input(process_type, arguments) = {
#	"type": process,
#	"command": concat(" ", array.concat(["kubelet"], arguments)),
#}
