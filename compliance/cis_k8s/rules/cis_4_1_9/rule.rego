package compliance.cis_k8s.rules.cis_4_1_9

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive (Automated)
finding = result {
	data_adapter.filename == "config.yaml"
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
	}
}

metadata = {
	"name": "Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive",
	"description": "Ensure that if the kubelet refers to a configuration file with the --config argument, that file has permissions of 644 or more restrictive.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.1.9", "Worker Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chmod 644 /var/lib/kubelet/config.yaml",
}
