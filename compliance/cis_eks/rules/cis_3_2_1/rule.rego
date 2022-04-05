package compliance.cis_eks.rules.cis_3_2_1

import data.compliance.cis_eks
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := cis_eks.data_adapter.process_args

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
	"tags": array.concat(cis_eks.default_tags, ["CIS 3.2.1", "Kubelet"]),
	"benchmark": cis_eks.benchmark_metadata,
	"default_value": "By default, anonymous access is enabled.",
	"remediation": `If modifying the Kubelet config file, edit the kubelet-config.json file /etc/kubernetes/kubelet/kubelet-config.json and set the below parameter to false.
"authentication": { "anonymous": { "enabled": false}}.
If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubelet-args.conf on each worker node and add the below parameter at the end of the KUBELET_ARGS variable string.
--anonymous-auth=false.
If using the api configz endpoint consider searching for the status of "authentication.*anonymous":{"enabled":false}" by extracting the live configuration from the nodes running kubelet.`,
}
