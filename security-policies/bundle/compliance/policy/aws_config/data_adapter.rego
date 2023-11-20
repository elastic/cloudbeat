package compliance.policy.aws_config.data_adapter

import future.keywords.if

is_configservice if {
	input.subType == "aws-config"
}

configs := input.resource
