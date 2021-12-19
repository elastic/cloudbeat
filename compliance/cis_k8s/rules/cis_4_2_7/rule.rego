package compliance.cis_k8s.rules.cis_4_2_7

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --make-iptables-util-chains argument is set to true (Automated)

# todo: If the --make-iptables-util-chains argument does not exist, and there is a Kubelet config file specified by --config, verify that the file does not set makeIPTablesUtilChains to false.
finding = result {
	# filter
	data_adapter.is_kubelet

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation = assert.is_false(common.contains_key_with_value(process_args, "--make-iptables-util-chains", "false"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --make-iptables-util-chains argument is set to true",
	"description": "Allow Kubelet to manage iptables.",
	"impact": "Kubelet would manage the iptables on the system and keep it in sync. If you are using any other iptables management solution, then there might be some conflicts.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.7", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "If using a Kubelet config file, edit the file to set makeIPTablesUtilChains: true. If using command line arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and remove the --make-iptables-util-chains argument from the KUBELET_SYSTEM_PODS_ARGS variable. Based on your system, restart the kubelet service.",
}
