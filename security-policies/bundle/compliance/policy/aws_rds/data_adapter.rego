package compliance.policy.aws_rds.data_adapter

is_rds {
	input.subType == "aws-rds"
}

storage_encrypted := input.resource.storage_encrypted

auto_minor_version_upgrade := input.resource.auto_minor_version_upgrade

publicly_accessible := input.resource.publicly_accessible

subnets := input.resource.subnets
