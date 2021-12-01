package compliance.cis_k8s.rules.cis_2_1

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated)
command_args := data_adapter.etcd_args

default rule_evaluation = false

rule_evaluation {
	command_args["--cert-file"]
	command_args["--key-file"]
}

finding = result {
	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --cert-file and --key-file arguments are set as appropriate",
	"description": "Configure TLS encryption for the etcd service.",
	"impact": "Client connections only over TLS would be served.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 2.1", "etcd"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Follow the etcd service documentation and configure TLS encryption. Then, edit the etcd pod specification file /etc/kubernetes/manifests/etcd.yaml on the master node and set  --cert-file=</path/to/ca-file> --key-file=</path/to/key-file>",
}
