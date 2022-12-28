package cis_aws.test_data

future_date = "2022-12-25T12:43:00+00:00"

past_date = "2021-12-25T12:43:00+00:00"

generate_password_policy(pwd_len, reuse_count) = {
	"resource": {
		"max_age_days": 90,
		"minimum_length": pwd_len,
		"require_lowercase": true,
		"require_numbers": true,
		"require_symbols": true,
		"require_uppercase": true,
		"reuse_prevention_count": reuse_count,
	},
	"type": "identity-management",
	"subType": "aws-password-policy",
}

not_evaluated_pwd_policy = {
	"type": "some type",
	"subType": "some sub type",
	"resource": {
		"max_age_days": 90,
		"minimum_length": 8,
		"require_lowercase": true,
		"require_numbers": true,
		"require_symbols": true,
		"require_uppercase": true,
		"reuse_prevention_count": 5,
	},
}

not_evaluated_iam_user = {
	"type": "identity-management",
	"subType": "aws-iam-user",
	"resource": {
		"name": "<root_account>",
		"access_keys": "test",
		"mfa_active": "test",
		"last_access": "test",
		"password_enabled": "test",
		"arn": "arn:aws:iam::704479110758:user/root",
	},
}

generate_iam_user(access_keys, mfa_active, has_logged_in, last_access, password_last_changed) = {
	"type": "identity-management",
	"subType": "aws-iam-user",
	"resource": {
		"name": "test",
		"access_keys": access_keys,
		"mfa_active": mfa_active,
		"last_access": last_access,
		"password_enabled": has_logged_in,
		"password_last_changed": password_last_changed,
		"arn": "arn:aws:iam::704479110758:user/test",
	},
}

generate_nacl(entry) = {
	"resource": {
		"Associations": [],
		"Entries": [entry],
		"IsDefault": false,
		"Tags": [],
	},
	"type": "ec2",
	"subType": "aws-nacl",
}
