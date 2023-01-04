package cis_aws.test_data

current_date := create_date_from_ns(time.now_ns())

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
	"subType": "gcp-iam-user",
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

generate_root_user(access_keys, mfa_active, last_access, mfa_devices) = {
	"type": "identity-management",
	"subType": "aws-iam-user",
	"resource": {
		"name": "<root_account>",
		"access_keys": access_keys,
		"mfa_active": mfa_active,
		"mfa_devices": mfa_devices,
		"last_access": last_access,
		"password_enabled": false,
		"password_last_changed": "not_supported",
		"arn": "arn:aws:iam::704479110758:root",
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

create_date_from_ns(x) = time_str {
	date := time.date(x)
	t := time.clock(x)

	time_str := sprintf("%d-%02d-%02dT%02d:%02d:%02d+00:00", array.concat(date, t))
}

not_evaluated_s3_bucket = {
	"resource": {
		"Name": "my-bucket",
		"SSEAlgorithm": "AES256",
	},
	"type": "wrong type",
	"subType": "wrong sub type",
}

generate_s3_bucket(name, sse_algorithm) = {
	"resource": {
		"Name": name,
		"SSEAlgorithm": sse_algorithm,
	},
	"type": "cloud-storage",
	"subType": "aws-s3",
}

generate_security_group(entry) = {
	"resource": entry,
	"type": "ec2",
	"subType": "aws-security-group",
}
