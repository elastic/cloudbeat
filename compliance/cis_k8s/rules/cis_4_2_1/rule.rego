package compliance.cis_k8s.rules.cis_4_2_1

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := cis_k8s.data_adapter.process_args

rule_evaluation {
	common.contains_key_with_value(process_args, "--anonymous-auth", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authentication:anonymous: enabled set to false.
rule_evaluation {
	not process_args["--anonymous-auth"]
	assert.is_false(data_adapter.process_config.config.authentication.anonymous.enabled)
}

# Ensure that the --anonymous-auth argument is set to false (Automated)
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
	"name": "Ensure that the --anonymous-auth argument is set to false",
	"description": "Disable anonymous requests to the Kubelet server.",
	"impact": "Anonymous requests will be rejected.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.1", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"default_value": "By default, anonymous access is enabled.",
	"remediation": "If using a Kubelet config file, edit the file to set authentication: anonymous: enabled to false. If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable. --anonymous-auth=false Based on your system, restart the kubelet service.",
}
