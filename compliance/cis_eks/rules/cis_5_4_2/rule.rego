package compliance.cis_eks.rules.cis_5_4_2

import data.compliance.cis_eks
import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

default rule_evaluation = false

# Allow only private access to cluster.
rule_evaluation {
	input.resource.Cluster.ResourcesVpcConfig.EndpointPrivateAccess
	not input.resource.Cluster.ResourcesVpcConfig.EndpointPublicAccess
}

# Ensure there Kuberenetes endpoint private access is enabled
finding = result {
	# filter
	data_adapter.is_aws_eks

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"endpoint_public_access": input.resource.Cluster.ResourcesVpcConfig.EndpointPublicAccess,
			"endpoint_private_access": input.resource.Cluster.ResourcesVpcConfig.EndpointPrivateAccess,
		},
	}
}

metadata = {
	"name": "Ensure clusters are created with Private Endpoint Enabled and Public Access Disabled",
	"description": "Disable access to the Kubernetes API from outside the node network if it is not required.",
	"rationale": `In a private cluster, the master node has two endpoints, a private and public endpoint.
The private endpoint is the internal IP address of the master, behind an internal load balancer in the master's VPC network.
Nodes communicate with the master using the private endpoint.
The public endpoint enables the Kubernetes API to be accessed from outside the master's VPC network.
Although Kubernetes API requires an authorized token to perform sensitive actions, a vulnerability could potentially expose the Kubernetes publically with unrestricted access.
Additionally, an attacker may be able to identify the current cluster and Kubernetes API version and determine whether it is vulnerable to an attack.
Unless required, disabling public endpoint will help prevent such threats, and require the attacker to be on the master's VPC network to perform any attack on the Kubernetes API.`,
	"remediation": ``,
	"tags": array.concat(cis_eks.default_tags, ["CIS 5.4.2", "Cluster Networking"]),
	"default_value": "By default, the Private Endpoint is disabled.",
	"benchmark": cis_eks.benchmark_metadata,
	"impact": `Configure the EKS cluster endpoint to be private. See Modifying Cluster Endpoint Access for further information on this topic.
1. Leave the cluster endpoint public and specify which CIDR blocks can communicate with the cluster endpoint.
The blocks are effectively a whitelisted set of public IP addresses that are allowed to access the cluster endpoint.
2. Configure public access with a set of whitelisted CIDR blocks and set private endpoint access to enabled.
This will allow public access from a specific range of public IPs while forcing all network traffic between the kubelets (workers) and the Kubernetes API through the cross-account ENIs that get provisioned into the cluster VPC when the control plane is provisioned.`,
}
