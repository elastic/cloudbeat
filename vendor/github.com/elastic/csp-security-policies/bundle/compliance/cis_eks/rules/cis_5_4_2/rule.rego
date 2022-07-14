package compliance.cis_eks.rules.cis_5_4_2

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
