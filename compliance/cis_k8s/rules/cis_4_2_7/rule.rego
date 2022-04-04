package compliance.cis_k8s.rules.cis_4_2_7

import data.compliance.cis_k8s
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
	"impact": "Kubelet would manage the iptables on the system and keep it in sync. If you are using any other iptables management solution, then there might be some conflicts.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.7", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "If using a Kubelet config file, edit the file to set makeIPTablesUtilChains: true. If using command line arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and remove the --make-iptables-util-chains argument from the KUBELET_SYSTEM_PODS_ARGS variable. Based on your system, restart the kubelet service.",
	"deafult_value": "By default, --make-iptables-util-chains argument is set to true.",
}
