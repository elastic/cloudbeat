package cis_k8s.test_data

# test data generater
filesystem_input(filename, mode, uid, gid) = {
	"type": "filesystem",
	"path": sprintf("file/path/%s", [filename]),
	"filename": filename,
	"mode": mode,
	"uid": uid,
	"gid": gid,
}
