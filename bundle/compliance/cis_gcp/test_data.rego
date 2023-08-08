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
	"subType": "gcp-iam-service-account",
}

generate_gcp_asset(type, subtype, resource, iam_policy) = {
	"resource": {
		"resource": resource,
		"iam_policy": iam_policy,
	},
	"type": type,
	"subType": subtype,
}

generate_monitoring_asset(log_metrics, alerts) = {
	"resource": {
		"log_metrics": log_metrics,
		"alerts": alerts,
	},
	"type": "monitoring",
	"subType": "gcp-monitoring",
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

generate_compute_resource(subType, info) = {
	"resource": {"resource": {"data": info}},
	"type": "cloud-compute",
	"subType": subType,
}

generate_iam_service_account_key(resourceData) = {
	"resource": {
		"resource": {"data": resourceData},
		"iam_policy": {},
	},
	"type": "kidentity-management",
	"subType": "gcp-iam-service-account-key",
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
