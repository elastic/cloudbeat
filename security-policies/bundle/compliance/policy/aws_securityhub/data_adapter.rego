package compliance.policy.aws_securityhub.data_adapter

is_securityhub_subType {
	input.subType == "aws-securityhub"
}

securityhub_resource = input.resource
