package compliance.policy.aws_kms.data_adapter

is_kms {
	input.subType == "aws-kms"
}

key_rotation_enabled := input.resource.key_rotation_enabled

key_metadata := input.resource.key_metadata
