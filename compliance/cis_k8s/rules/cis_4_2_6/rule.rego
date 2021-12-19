package compliance.cis_k8s.rules.cis_4_2_6

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --protect-kernel-defaults argument is set to true (Automated)
# todo: If the --protect-kernel-defaults argument is not present, check that there is a Kubelet config file specified by --config, and that the file sets protectKernelDefaults to true.
finding = result {
	# filter
	data_adapter.is_kubelet

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation = common.contains_key_with_value(process_args, "--protect-kernel-defaults", "true")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --protect-kernel-defaults argument is set to true",
	"description": "Protect tuned kernel parameters from overriding kubelet default kernel parameter values.",
	"impact": "You would have to re-tune kernel parameters to match kubelet parameters.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 4.2.6", "Kubelet"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "If using a Kubelet config file, edit the file to set protectKernelDefaults: true. If using command line arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubeadm.conf on each worker node and set the below parameter in KUBELET_SYSTEM_PODS_ARGS variable. --protect-kernel-defaults=true Based on your system, restart the kubelet service.",
}
