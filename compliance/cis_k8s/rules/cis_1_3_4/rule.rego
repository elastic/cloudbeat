package compliance.cis_k8s.rules.cis_1_3_4

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --service-account-private-key-file argument is set as appropriate (Automated)
finding = result {
	# filter
	data_adapter.is_kube_controller_manger

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.contains_key(process_args, "--service-account-private-key-file")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --service-account-private-key-file argument is set as appropriate",
	"description": "To ensure that keys for service account tokens can be rotated as needed, a separate public/private key pair should be used for signing service account tokens. The private key should be specified to the controller manager with --service-account-private-key-file as appropriate.",
	"impact": "You would need to securely maintain the key file and rotate the keys based on your organization's key rotation policy.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.3.4", "Controller Manager"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Edit the Controller Manager pod specification file /etc/kubernetes/manifests/kube-controller-manager.yaml on the master node and set the --service-account-private-key-file parameter to the private key file for service accounts.",
}
