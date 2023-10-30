package compliance.policy.aws_eks.data_adapter

is_aws_eks {
	input.subType == "aws-eks"
}

cluster = input.resource.Cluster
