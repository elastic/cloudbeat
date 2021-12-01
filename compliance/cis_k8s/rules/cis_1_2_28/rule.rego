package compliance.cis_k8s.rules.cis_1_2_28

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --service-account-key-file argument is set as appropriate (Automated)
finding = result {
	command_args := data_adapter.api_server_command_args
	rule_evaluation := common.contains_key(command_args, "--service-account-key-file")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --service-account-key-file argument is set as appropriate",
	"description": "By default, if no --service-account-key-file is specified to the apiserver, it uses the private key from the TLS serving certificate to verify service account tokens. To ensure that the keys for service account tokens could be rotated as needed, a separate public/private key pair should be used for signing service account tokens. Hence, the public key should be specified to the apiserver with --service-account-key-file.",
	"impact": "The corresponding private key must be provided to the controller manager. You would need to securely maintain the key file and rotate the keys based on your organization's key rotation policy.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.28", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the below parameter. --service-account-lookup=true",
}
