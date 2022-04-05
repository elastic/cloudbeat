package compliance.cis_eks.rules.cis_3_2_3

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := cis_eks.data_adapter.process_args

rule_evaluation {
	common.contains_key(process_args, "--client-ca-file")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authentication:x509:clientCAFile: set to a valid path.
rule_evaluation {
	data_adapter.process_config.config.authentication.x509.clientCAFile
}

# Ensure that the --client-ca-file argument is set as appropriate (Automated)
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
	"name": "Ensure that the --client-ca-file argument is set as appropriate",
	"description": "Enable Kubelet authentication using certificates.",
	"impact": "You require TLS to be configured on apiserver as well as kubelets.",
	"tags": array.concat(cis_eks.default_tags, ["CIS 3.2.3", "Kubelet"]),
	"benchmark": cis_eks.benchmark_metadata,
	"remediation": `If modifying the Kubelet config file, edit the kubelet-config.json file /etc/kubernetes/kubelet/kubelet-config.json and set the below parameter to false
"authentication": { "x509": {"clientCAFile:" to the location of the client CA file}}.
If using executable arguments, edit the kubelet service file /etc/systemd/system/kubelet.service.d/10-kubelet-args.conf on each worker node and add the below parameter at the end of the KUBELET_ARGS variable string. 
--client-ca-file=<path/to/client-ca-file>
If using the api configz endpoint consider searching for the status of "authentication.*x509":("clientCAFile":"/etc/kubernetes/pki/ca.crt" by extracting the live configuration from the nodes running kubelet.`,
	"rationale": `The connections from the apiserver to the kubelet are used for fetching logs for pods, attaching (through kubectl) to running pods, and using the kubelet’s port-forwarding functionality.
These connections terminate at the kubelet’s HTTPS endpoint.
By default, the apiserver does not verify the kubelet’s serving certificate, which makes the connection subject to man-in-the-middle attacks, and unsafe to run over untrusted and/or public networks. Enabling Kubelet certificate authentication ensures that the apiserver could authenticate the Kubelet before submitting any requests.`,
	"default_value": "See the EKS documentation for the default value.",
}
