package compliance.cis_k8s.rules.cis_1_1_19

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the Kubernetes PKI directory and file ownership is set to root:root
finding = result {
	common.file_in_path("/etc/kubernetes/pki/", data_adapter.file_path)

	uid = data_adapter.owner_user_id
	gid = data_adapter.owner_group_id
	rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"uid": "root", "gid": "root"},
		"evidence": {"uid": uid, "gid": gid},
	}
}

metadata = {
	"name": "Ensure that the Kubernetes PKI directory and file ownership is set to root:root",
	"description": "Kubernetes makes use of a number of certificates as part of its operation. You should set the ownership of the directory containing the PKI information and all files in that directory to maintain their integrity. The directory and files should be owned by root:root.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.19", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chown -R root:root /etc/kubernetes/pki/",
}
