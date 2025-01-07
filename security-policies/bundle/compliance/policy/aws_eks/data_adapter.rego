package compliance.policy.aws_eks.data_adapter

import future.keywords.if

is_aws_eks if {
	input.subType == "aws-eks"
}

cluster := input.resource.Cluster
