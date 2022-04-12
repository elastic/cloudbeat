package compliance.cis_k8s.rules.cis_1_1_17

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive (Automated)
finding = result {
	data_adapter.filename == "controller-manager.conf"
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"filemode": 644},
		"evidence": {"filemode": filemode},
	}
}

metadata = {
	"name": "Ensure that the controller-manager.conf file has permissions of 644 or more restrictive.",
	"description": "The controller-manager.conf file is the kubeconfig file for the Controller Manager. You should restrict its file permissions to maintain the integrity of the file. The file should be writable by only the administrators on the system.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.17", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chmod 644 /etc/kubernetes/controller-manager.conf",
}
