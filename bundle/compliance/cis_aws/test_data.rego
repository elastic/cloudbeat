package cis_aws.test_data

generate_password_policy(pwd_len, reuse_count) = {
	"resource": {
		"MaxAgeDays": 90,
		"MinimumLength": pwd_len,
		"RequireLowercase": true,
		"RequireNumbers": true,
		"RequireSymbols": true,
		"RequireUppercase": true,
		"ReusePreventionCount": reuse_count,
	},
	"type": "identity-management",
	"subType": "aws-password-policy",
}

not_evaluated_input = {
	"type": "some type",
	"subType": "some sub type",
	"resource": {
		"MaxAgeDays": 90,
		"MinimumLength": 8,
		"RequireLowercase": true,
		"RequireNumbers": true,
		"RequireSymbols": true,
		"RequireUppercase": true,
		"ReusePreventionCount": 5,
	},
}
