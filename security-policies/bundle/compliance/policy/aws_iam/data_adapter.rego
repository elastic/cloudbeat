package compliance.policy.aws_iam.data_adapter

is_server_certificate {
	input.subType == "aws-iam-server-certificate"
}

is_pwd_policy {
	input.subType == "aws-password-policy"
}

is_iam_user {
	input.subType == "aws-iam-user"
	input.resource.name != "<root_account>"
}

is_root_user {
	input.subType == "aws-iam-user"
	input.resource.name == "<root_account>"
}

is_iam_policy {
	input.subType == "aws-policy"
}

is_aws_support_access {
	is_iam_policy
	input.resource.Arn == "arn:aws:iam::aws:policy/AWSSupportAccess"
}

is_access_analyzers {
	input.subType == "aws-access-analyzers"
}

pwd_policy = policy {
	is_pwd_policy
	policy := input.resource
}

iam_user = input.resource

policy_document = input.resource.document

roles = input.resource.roles

server_certificates = input.resource.certificates

analyzers = input.resource.Analyzers

analyzer_regions = input.resource.Regions

used_active_access_keys = {access_key |
	access_key = iam_user.access_keys[_]
	access_key.active
	access_key.has_used
}

unused_active_access_keys = {access_key |
	access_key = iam_user.access_keys[_]
	access_key.active
	not access_key.has_used
}

active_access_keys := used_active_access_keys | unused_active_access_keys
