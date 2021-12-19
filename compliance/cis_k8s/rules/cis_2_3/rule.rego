package compliance.cis_k8s.rules.cis_2_3

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --auto-tls argument is not set to true (Automated)
finding = result {
	# filter
	data_adapter.is_etcd

	# Verify that if the --auto-tls argument exists, it is not set to true.

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--auto-tls", "true") == false

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --auto-tls argument is not set to true",
	"description": "Do not use self-signed certificates for TLS.",
	"impact": "Clients will not be able to use self-signed certificates for TLS.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 2.3", "etcd"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the etcd pod specification file /etc/kubernetes/manifests/etcd.yaml on the master node and either remove the --auto-tls parameter or set it to false. --auto-tls=false",
}
