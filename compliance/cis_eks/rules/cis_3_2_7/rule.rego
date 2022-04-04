package compliance.cis_eks.rules.cis_3_2_7

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := data_adapter.process_args

# Ensure that the --make-iptables-util-chains argument is set to true (Automated)
rule_evaluation {
	common.contains_key_with_value(process_args, "--make-iptables-util-chains", "true")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for makeIPTablesUtilChains is set to true.
rule_evaluation {
	not process_args["--make-iptables-util-chains"]
	data_adapter.process_config.config.makeIPTablesUtilChains
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
	"name": "Ensure that the --make-iptables-util-chains argument is set to true",
	"description": "Allow Kubelet to manage iptables.",
	"impact": `Kubelet would manage the iptables on the system and keep it in sync.
If you are using any other iptables management solution, then there might be some conflicts.`,
	"tags": array.concat(cis_eks.default_tags, ["CIS 3.2.7", "Kubelet"]),
	"benchmark": cis_eks.benchmark_metadata,
	"remediation": `If modifying the Kubelet config file, edit the kubelet-config.json file /etc/kubernetes/kubelet/kubelet-config.json and set the below parameter to false
"makeIPTablesUtilChains": true
If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubelet-args.conf on each worker node and add the below parameter at the end of the KUBELET_ARGS variable string.
--make-iptables-util-chains:true
If using the api configz endpoint consider searching for the status of "makeIPTablesUtilChains": true by extracting the live configuration from the nodes running kubelet.`,
	"default_value": "See the Amazon EKS documentation for the default value.",
	"rationale": `Kubelets can automatically manage the required changes to iptables based on how you choose your networking options for the pods.
It is recommended to let kubelets manage the changes to iptables.
This ensures that the iptables configuration remains in sync with pods networking configuration.
Manually configuring iptables with dynamic pod network configuration changes might hamper the communication between pods/containers and to the outside world.
You might have iptables rules too restrictive or too open.`,
}
