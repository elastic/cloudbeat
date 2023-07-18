package cis_gcp.test_data

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
	"subType": "gcp-kms",
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
	"subType": "gcp-gcs",
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
