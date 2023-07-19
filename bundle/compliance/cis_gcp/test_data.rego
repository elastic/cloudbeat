package cis_gcp.test_data

generate_iam_policy(members, role) = {
	"resource": {
		"resource": {},
		"iam_policy": {"bindings": [{
			"role": role,
			"members": members,
		}]},
	},
	"type": "key-management",
	"subType": "gcp-iam",
}

generate_kms_resource(members, rotationPeriod, nextRotationTime, primary) = {
	"resource": {
		"resource": {"data": {
			"nextRotationTime": nextRotationTime,
			"rotationPeriod": rotationPeriod,
			"primary": primary,
		}},
		"iam_policy": {"bindings": [{
			"role": "roles/cloudkms.cryptoKeyEncrypterDecrypter",
			"members": members,
		}]},
	},
	"type": "key-management",
	"subType": "gcp-cloudkms-crypto-key",
}

generate_gcs_resource(members, isBucketLevelAccessEnabled) = {
	"resource": {
		"resource": {"data": {"iamConfiguration": {"uniformBucketLevelAccess": {"enabled": isBucketLevelAccessEnabled}}}},
		"iam_policy": {"bindings": [{
			"role": "roles/storage.objectViewer",
			"members": members,
		}]},
	},
	"type": "cloud-storage",
	"subType": "gcp-storage-bucket",
}

generate_bq_resource(config, subType, members) = {
	"resource": {
		"resource": {"data": {"defaultEncryptionConfiguration": config}},
		"iam_policy": {"bindings": [{
			"role": "roles/bigquery.dataViewer",
			"members": members,
		}]},
	},
	"type": "cloud-storage",
	"subType": subType,
}

not_eval_resource = {
	"resource": {},
	"type": "key-management",
	"subType": "no-exisitng-type",
}

# missing resource.iam_policy
no_policy_resource = {
	"resource": {"resource": {}},
	"type": "key-management",
	"subType": "gcp-iam",
}
