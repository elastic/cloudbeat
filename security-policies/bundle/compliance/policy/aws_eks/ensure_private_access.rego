package compliance.policy.aws_eks.ensure_private_access

import data.compliance.lib.common
import data.compliance.policy.aws_eks.data_adapter
import future.keywords.if

# Allow only private access to cluster.
is_only_private(cluster, cidr_allowed) if {
	cluster.ResourcesVpcConfig.EndpointPrivateAccess
	public_access_is_restricted(cluster, cidr_allowed)
} else := false

public_access_is_restricted(cluster, _) if {
	not cluster.ResourcesVpcConfig.EndpointPublicAccess
}

public_access_is_restricted(cluster, cidr_allowed) if {
	cidr_allowed == true

	cluster.ResourcesVpcConfig.EndpointPublicAccess
	public_access_cidrs := cluster.ResourcesVpcConfig.PublicAccessCidrs

	# Ensure that publicAccessCidr has a valid filter
	allow_all_filter := "0.0.0.0/0"
	invalid_filters := [index | public_access_cidrs[index] == allow_all_filter]
	count(invalid_filters) == 0
}

# Ensure there Kuberenetes endpoint private access is enabled
finding(cidr_allowed) := result if {
	# filter
	data_adapter.is_aws_eks

	cluster := data_adapter.cluster
	rule_evaluation := is_only_private(cluster, cidr_allowed)

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		object.union_n([
			{
				"endpoint_public_access": cluster.ResourcesVpcConfig.EndpointPublicAccess,
				"endpoint_private_access": cluster.ResourcesVpcConfig.EndpointPrivateAccess,
			},
			cidr_evidence(cluster.ResourcesVpcConfig, cidr_allowed),
		]),
	)
}

cidr_evidence(config, cidr_allowed) := result if {
	cidr_allowed == true
	result := {"public_access_cidrs": config.PublicAccessCidrs}
} else := {}
