package compliance.cis_eks.rules.cis_3_1_1

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubeconfig file permissions are set to 644 or more restrictive
finding = result {
	# filter
	data_adapter.filename == "kubeconfig"

	# evaluate
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
	}
}

metadata = {
	"name": "Ensure that the kubeconfig file permissions are set to 644 or more restrictive",
	"description": "If kubelet is running, and if it is using a file-based kubeconfig file, ensure that the proxy kubeconfig file has permissions of 644 or more restrictive.",
	"rationale": `The kubelet kubeconfig file controls various parameters of the kubelet service in the worker node.
You should restrict its file permissions to maintain the integrity of the file.
The file should be writable by only the administrators on the system.
It is possible to run kubelet with the kubeconfig parameters configured as a Kubernetes ConfigMap instead of a file. In this case, there is no proxy kubeconfig file.`,
	"impact": "None",
	"tags": array.concat(cis_eks.default_tags, ["CIS 3.1.1", "Worker Node Configuration"]),
	"default_value": "See the AWS EKS documentation for the default value.",
	"benchmark": cis_eks.benchmark_name,
	"remediation": "chmod 644 /var/lib/kubelet/kubeconfig",
}
