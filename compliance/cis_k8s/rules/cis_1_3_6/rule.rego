package compliance.cis_k8s.rules.cis_1_3_6

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the RotateKubeletServerCertificate argument is set to true (Automated)
finding = result {
	# filter
	data_adapter.is_kube_controller_manger

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--feature-gates", "RotateKubeletServerCertificate=true")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the RotateKubeletServerCertificate argument is set to true",
	"description": "Enable kubelet server certificate rotation on controller-manager.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.3.6", "Controller Manager"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the Controller Manager pod specification file /etc/kubernetes/manifests/kube-controller-manager.yaml on the master node and set the --feature-gates parameter to include RotateKubeletServerCertificate=true. --feature-gates=RotateKubeletServerCertificate=true",
}
