package compliance.cis_eks.rules.cis_3_2_2

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
# If the --authorization-mode argument is present check that it is not set to AlwaysAllow.

default rule_evaluation = false
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
	"tags": array.concat(cis_eks.default_tags, ["CIS 3.2.2", "Kubelet"]),
	"benchmark": cis_eks.benchmark_metadata,
	"remediation": `If modifying the Kubelet config file, edit the kubelet-config.json file /etc/kubernetes/kubelet/kubelet-config.json and set the below parameter to false.
"authentication"... "webhook":{"enabled":true...
If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubelet-args.conf on each worker node and add the below parameter at the end of the KUBELET_ARGS variable string.
--authorization-mode=Webhook.
If using the api configz endpoint consider searching for the status of "authentication.*webhook":{"enabled":true" by extracting the live configuration from the nodes running kubelet.`,
	"rationale": `Kubelets, by default, allow all authenticated requests (even anonymous ones) without needing explicit authorization checks from the apiserver.
You should restrict this behavior and only allow explicitly authorized requests.`,
	"default_value": "See the EKS documentation for the default value.",
}
