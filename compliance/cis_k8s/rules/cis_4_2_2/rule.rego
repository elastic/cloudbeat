package compliance.cis_k8s.rules.cis_4_2_2

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)

# If the --authorization-mode argument is present check that it is not set to AlwaysAllow.

# evaluate
process_args := data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--authorization-mode"]
	common.arg_values_contains(process_args, "--authorization-mode", "AlwaysAllow") == false
}

# todo: If it is not present check that there is a Kubelet config file specified by --config, and that file sets authorization: mode to something other than AlwaysAllow.

finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --authorization-mode argument is not set to AlwaysAllow",
	"description": "Do not allow all requests. Enable explicit authorization.",
	"impact": "Unauthorized requests will be denied.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.2", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "If using a Kubelet config file, edit the file to set authorization: mode to Webhook. If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and set the below parameter in KUBELET_AUTHZ_ARGS variable. --authorization-mode=Webhook Based on your system, restart the kubelet service",
}
