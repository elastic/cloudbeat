package compliance.cis_k8s.rules.cis_2_4

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated)

# evaluate
process_args := data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--peer-cert-file"]
	process_args["--peer-key-file"]
}

finding = result {
	# filter
	data_adapter.is_etcd

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate",
	"description": "etcd should be configured to make use of TLS encryption for peer connections.",
	"impact": "etcd cluster peers would need to set up TLS for their communication.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 2.4", "etcd"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Follow the etcd service documentation and configure peer TLS encryption as appropriate for your etcd cluster. Then, edit the etcd pod specification file /etc/kubernetes/manifests/etcd.yaml on the master node and set --peer-cert-file=</path/to/peer-cert-file> --peer-key-file=</path/to/peer-key-file>",
}
