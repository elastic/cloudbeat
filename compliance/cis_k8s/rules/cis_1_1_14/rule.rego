package compliance.cis_k8s.rules.cis_1_1_14

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admin.conf file ownership is set to root:root (Automated)
finding = result {
	data_adapter.filename == "admin.conf"
	uid = data_adapter.owner_user_id
	gid = data_adapter.owner_group_id
	rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"uid": uid, "gid": gid},
	}
}

metadata = {
	"name": "Ensure that the API server pod specification file ownership is set to root:root",
	"description": "The admin.conf file contains the admin credentials for the cluster. You should set its file ownership to maintain the integrity of the file. The file should be owned by root:root.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.14", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chown root:root /etc/kubernetes/admin.conf",
}
