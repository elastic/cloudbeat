package compliance.policy.aws_securityhub.data_adapter

import future.keywords.if

is_securityhub_subType if {
	input.subType == "aws-securityhub"
}

securityhub_resource := input.resource
