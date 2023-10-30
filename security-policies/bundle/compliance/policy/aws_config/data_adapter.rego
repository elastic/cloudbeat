package compliance.policy.aws_config.data_adapter

is_configservice {
	input.subType == "aws-config"
}

configs := input.resource
