package compliance.cis_k8s.rules.cis_4_2_3

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --client-ca-file argument is set as appropriate (Automated)
finding = result {
	# filter
	data_adapter.is_kubelet

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation = common.contains_key(process_args, "--client-ca-file")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --client-ca-file argument is set as appropriate",
	"description": "Enable Kubelet authentication using certificates.",
	"impact": "You require TLS to be configured on apiserver as well as kubelets.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.3", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "If using a Kubelet config file, edit the file to set authentication: x509: clientCAFile to the location of the client CA file. If using command line arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and set the below parameter in KUBELET_AUTHZ_ARGS variable. --client-ca-file=<path/to/client-ca-file> Based on your system, restart the kubelet service.",
}
