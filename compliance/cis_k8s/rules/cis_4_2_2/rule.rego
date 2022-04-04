package compliance.cis_k8s.rules.cis_4_2_2

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
# If the --authorization-mode argument is present check that it is not set to AlwaysAllow.

default rule_evaluation = false

# evaluate
process_args := data_adapter.process_args

rule_evaluation {
	is_authorization_allow_all
}

is_authorization_allow_all {
	process_args["--authorization-mode"]
	not common.contains_key_with_value(process_args, "--authorization-mode", "AlwaysAllow")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authorization:mode is not set to AlwaysAllow.
rule_evaluation {
	not is_authorization_allow_all
	data_adapter.process_config.config.authorization.mode
	not data_adapter.process_config.config.authorization.mode == "AlwaysAllow"
}

finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"process_args": process_args,
			"process_config": data_adapter.process_config,
		},
	}
}

metadata = {
	"name": "Ensure that the --authorization-mode argument is not set to AlwaysAllow",
	"description": "Do not allow all requests. Enable explicit authorization.",
	"impact": "Unauthorized requests will be denied.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.2", "Kubelet"]),
	"default_value": "By default, --authorization-mode argument is set to AlwaysAllow.",
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "If using a Kubelet config file, edit the file to set authorization: mode to Webhook. If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and set the below parameter in KUBELET_AUTHZ_ARGS variable. --authorization-mode=Webhook Based on your system, restart the kubelet service",
}
