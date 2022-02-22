package compliance.cis_k8s.rules.cis_1_2_29

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated)

# evaluate
process_args := data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--etcd-certfile"]
	process_args["--etcd-keyfile"]
}

finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate",
	"description": "etcd is a highly-available key value store used by Kubernetes deployments for persistent storage of all of its REST API objects. These objects are sensitive in nature and should be protected by client authentication. This requires the API server to identify itself to the etcd server using a client certificate and key.",
	"impact": "TLS and client certificate authentication must be configured for etcd.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.29", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Follow the Kubernetes documentation and set up the TLS connection between the apiserver and etcd. Then, edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the etcd certificate and key file parameters.",
}
