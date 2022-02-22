package compliance.cis_k8s.rules.cis_4_1_10

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubelet --config configuration file ownership is set to root:root (Automated)
finding = result {
	data_adapter.filename == "kubelet.conf"
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
	"name": "Ensure that the kubelet --config configuration file ownership is set to root:root",
	"description": "Ensure that if the kubelet refers to a configuration file with the --config argument, that file is owned by root:root",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.1.10", "Worker Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chown root:root /etc/kubernetes/kubelet.conf",
}
