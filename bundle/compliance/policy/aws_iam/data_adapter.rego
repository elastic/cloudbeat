package compliance.policy.aws_iam.data_adapter

is_pwd_policy {
	input.subType == "aws-password-policy"
}

pwd_policy = policy {
	policy := input.resource
}
