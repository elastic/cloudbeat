package cis_k8s.test_data

# test data generater

# genrates `file-system` type input data
filesystem_input(filename, mode, uid, gid) = {"resource": {
	"type": "file-system",
	"path": sprintf("file/path/%s", [filename]),
	"filename": filename,
	"mode": mode,
	"uid": uid,
	"gid": gid,
}}

# genrates `process` type input data
process_input(process_name, arguments) = {"resource": {
	"type": "process",
	"command": concat(" ", array.concat([process_name], arguments)),
}}

kube_api_input(resource) = {"resource": object.union({"type": "kube-api"}, resource)}
