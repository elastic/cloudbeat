package cis_aws.test_data

generate_password_policy(pwd_len, reuse_count) := {
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

not_evaluated_pwd_policy := {
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

not_evaluated_iam_user := {
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

generate_iam_user(access_keys, mfa_active, has_logged_in, last_access, password_last_changed) := {
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

generate_iam_user_with_policies(inline_policies, attached_policies) := {
	"type": "identity-management",
	"subType": "aws-iam-user",
	"resource": {
		"name": "test",
		"inline_policies": inline_policies,
		"attached_policies": attached_policies,
	},
}

generate_root_user(access_keys, mfa_active, last_access, mfa_devices) := {
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

generate_nacl(entry) := {
	"resource": {
		"Associations": [],
		"Entries": [entry],
		"IsDefault": false,
		"Tags": [],
	},
	"type": "ec2",
	"subType": "aws-nacl",
}

not_evaluated_s3_bucket := {
	"resource": {
		"name": "my-bucket",
		"sse_algorithm": "AES256",
		"bucket_policy": {
			"Version": "2012-10-17",
			"Statement": [generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "true")],
		},
		"bucket_versioning": generate_s3_bucket_versioning(true, true),
		"public_access_block_configuration": generate_s3_public_access_block_configuration(true, true, true, true),
		"account_public_access_block_configuration": generate_s3_public_access_block_configuration(true, true, true, true),
	},
	"type": "wrong type",
	"subType": "wrong sub type",
}

generate_s3_bucket(name, sse_algorithm, bucket_policy_statement, bucket_versioning, public_access_block_configuration, account_public_access_block_configuration) := {
	"resource": {
		"name": name,
		"sse_algorithm": sse_algorithm,
		"bucket_policy": {
			"Version": "1",
			"Statement": bucket_policy_statement,
		},
		"bucket_versioning": bucket_versioning,
		"public_access_block_configuration": public_access_block_configuration,
		"account_public_access_block_configuration": account_public_access_block_configuration,
	},
	"type": "cloud-storage",
	"subType": "aws-s3",
}

generate_s3_bucket_policy_statement(effect, principal, action, is_secure_transport) := {
	"Sid": "Statement1",
	"Effect": effect,
	"Principal": principal,
	"Action": action,
	"Resource": "arn:aws:s3:::test-bucket/*",
	"Condition": {"Bool": {"aws:SecureTransport": is_secure_transport}},
}

generate_s3_bucket_versioning(enabled, mfa_delete) := {
	"Enabled": enabled,
	"MfaDelete": mfa_delete,
}

s3_bucket_without_policy := {
	"resource": {
		"name": "my-bucket",
		"sse_algorithm": "AES256",
		"bucket_versioning": "",
	},
	"type": "cloud-storage",
	"subType": "aws-s3",
}

generate_security_group(entry) := {
	"resource": entry,
	"type": "ec2",
	"subType": "aws-security-group",
}

generate_monitoring_resources(items) := {
	"resource": {"Items": items},
	"type": "monitoring",
	"subType": "aws-multi-trails",
}

generate_securityhub(sb) := {
	"resource": sb,
	"type": "monitoring",
	"subType": "aws-securityhub",
}

generate_enriched_trail(is_log_validation_enabled, cloudwatch_log_group_arn, log_delivery_time, is_bucket_logging_enabled, kms_key_id) := {
	"type": "cloud-audit",
	"subType": "aws-trail",
	"resource": {
		"Trail": {
			"LogFileValidationEnabled": is_log_validation_enabled,
			"CloudWatchLogsLogGroupArn": cloudwatch_log_group_arn,
			"KmsKeyId": kms_key_id,
		},
		"Status": {"LatestCloudWatchLogsDeliveryTime": log_delivery_time},
		"bucket_info": {"logging": {"Enabled": is_bucket_logging_enabled}},
	},
}

create_bucket_acl(principal_uri) := {
	"Owner": {
		"ID": "f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8",
		"DisplayName": "exampleuser",
	},
	"Grants": [
		{
			"Grantee": {
				"Type": "CanonicalUser",
				"ID": "f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8",
				"DisplayName": "exampleuser",
			},
			"Permission": "FULL_CONTROL",
		},
		{
			"Grantee": {
				"Type": "Group",
				"ID": "f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8f5c5b99a8",
				"DisplayName": "exampleuser",
				"URI": principal_uri,
			},
			"Permission": "FULL_CONTROL",
		},
	],
}

generate_trail_bucket_info(principal_uri, policy_statements) := {
	"type": "cloud-audit",
	"subType": "aws-trail",
	"resource": {"bucket_info": {"acl": create_bucket_acl(principal_uri), "policy": {"Version": "2012-10-17", "Statement": policy_statements}}},
}

generate_event_selectors(entries, is_multi_region) := {
	"type": "cloud-audit",
	"subType": "aws-trail",
	"resource": {"Trail": {"IsMultiRegionTrail": is_multi_region}, "EventSelectors": entries},
}

generate_vpc_resource(flow_logs) := {
	"resource": {"flow_logs": flow_logs},
	"type": "ec2",
	"subType": "aws-vpc",
}

generate_ebs_encryption_resource(encryption_enabled) := {
	"resource": {"enabled": encryption_enabled},
	"type": "cloud-compute",
	"subType": "aws-ebs",
}

not_evaluated_trail := {
	"type": "cloud-audit",
	"subType": "not-an-aws-trail",
	"resource": {"log_file_validation_enabled": false},
}

not_evaluated_rds_db_instance := {
	"resource": {
		"identifier": "test-db",
		"arn": "arn:aws:rds:eu-west-1:704479110758:db:devops-postgres-rds",
		"storage_encrypted": true,
		"auto_minor_version_upgrade": true,
		"publicly_accessible": true,
		"subnets": [generate_rds_db_instance_subnet_with_route("0.0.0.0/0", "igw-a25733d9")],
	},
	"type": "wrong type",
	"subType": "wrong sub type",
}

generate_rds_db_instance(encryption_enabled, auto_minor_version_upgrade_enabled, publicly_accessible, subnets) := {
	"resource": {
		"identifier": "test-db",
		"arn": "arn:aws:rds:eu-west-1:704479110758:db:devops-postgres-rds",
		"storage_encrypted": encryption_enabled,
		"auto_minor_version_upgrade": auto_minor_version_upgrade_enabled,
		"publicly_accessible": publicly_accessible,
		"subnets": subnets,
	},
	"type": "cloud-database",
	"subType": "aws-rds",
}

generate_rds_db_instance_subnet_with_route(destination_cidr_block, gateway_id) := {
	"ID": "subnet-12345678",
	"RouteTable": {
		"ID": "rtb-12345678",
		"Routes": [{
			"DestinationCidrBlock": destination_cidr_block,
			"GatewayId": gateway_id,
		}],
	},
}

generate_s3_public_access_block_configuration(block_public_acls, block_public_policy, ignore_public_acls, restrict_public_buckets) := {
	"BlockPublicAcls": block_public_acls,
	"BlockPublicPolicy": block_public_policy,
	"IgnorePublicAcls": ignore_public_acls,
	"RestrictPublicBuckets": restrict_public_buckets,
}

generate_kms_resource(symmetric_default_enabled) := {
	"resource": {
		"key_metadata": {
			# Only relevent keys are included
			"KeyId": "21c0ba99-3a6c-4f72-8ef8-8118d4804710",
			"KeySpec": "SYMMETRIC_DEFAULT",
		},
		"key_rotation_enabled": symmetric_default_enabled,
	},
	"type": "key-management",
	"subType": "aws-kms",
}

generate_aws_configservice_with_resource(resource) := {
	"resource": resource,
	"type": "cloud-config",
	"subType": "aws-config",
}

generate_aws_configservice_recorder(all_supported_enabled, include_global_resource_types_enabled) := {"ConfigurationRecorder": {"RecordingGroup": {
	"AllSupported": all_supported_enabled,
	"IncludeGlobalResourceTypes": include_global_resource_types_enabled,
}}}

aws_configservice_disabled_region_recorder := generate_aws_configservice_with_resource([
	{"recorders": [
		generate_aws_configservice_recorder(true, true),
		generate_aws_configservice_recorder(false, false),
	]},
	{"recorders": [
		generate_aws_configservice_recorder(false, false),
		generate_aws_configservice_recorder(false, false),
	]},
])

aws_configservice_empty_recorders := generate_aws_configservice_with_resource([{"recorders": []}])
