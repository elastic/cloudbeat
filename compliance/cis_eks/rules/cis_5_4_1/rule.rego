package compliance.cis_eks.rules.cis_5_4_1

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

default rule_evaluation = false

# Verify that private access is enabled
# Verify that public access is enabled
# Restrict the public access to the cluster's control plane to only an allowlist of authorized IPs.
rule_evaluation {
	input.resource.Cluster.ResourcesVpcConfig.EndpointPrivateAccess
	public_access_is_restricted
}

public_access_is_restricted {
	not input.resource.Cluster.ResourcesVpcConfig.EndpointPublicAccess
}

public_access_is_restricted {
	input.resource.Cluster.ResourcesVpcConfig.EndpointPublicAccess
	public_access_cidrs := input.resource.Cluster.ResourcesVpcConfig.PublicAccessCidrs

	# Ensure that publicAccessCidr has a valid filter
	allow_all_filter := "0.0.0.0/0"
	invalid_filters := [index | public_access_cidrs[index] == allow_all_filter]
	count(invalid_filters) == 0
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
			"public_access_cidrs": input.resource.Cluster.ResourcesVpcConfig.PublicAccessCidrs,
		},
	}
}
