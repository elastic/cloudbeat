package compliance.cis_eks.rules.cis_5_4_3

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = true

# Verify that the node doesn't have an external IP
rule_evaluation = false {
	some address
	input.resource.status.addresses[address].type == "ExternalIP"
	input.resource.status.addresses[address].address != "0.0.0.0"
}

evidence["external_ip"] = result {
	not rule_evaluation
	input.resource.status.addresses[address].type == "ExternalIP"
	result = input.resource.status.addresses[address]
}

# Ensure there cluster node don't have a public IP
finding = result {
	# filter
	data_adapter.is_kube_node

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": evidence,
	}
}

metadata = {
	"name": "Ensure clusters are created with Private Nodes",
	"description": `Disable public IP addresses for cluster nodes, so that they only have private IP addresses.
Private Nodes are nodes with no public IP addresses.`,
	"rationale": `Disabling public IP addresses on cluster nodes restricts access to only internal networks, forcing attackers to obtain local network access before attempting to compromise the underlying Kubernetes hosts.`,
	"remediation": "",
	"tags": array.concat(cis_eks.default_tags, ["CIS 5.4.3", "Cluster Networking"]),
	"default_value": "By default, Private Nodes are disabled.",
	"benchmark": cis_eks.benchmark_metadata,
	"impact": `To enable Private Nodes, the cluster has to also be configured with a private master IP range and IP Aliasing enabled.
Private Nodes do not have outbound access to the public internet.
If you want to provide outbound Internet access for your private nodes, you can use Cloud NAT or you can manage your own NAT gateway.`,
}
