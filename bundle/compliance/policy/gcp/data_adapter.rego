package compliance.policy.gcp.data_adapter

resource = input.resource.asset.resource

iam_policy = object.get(input.resource.asset, "iam_policy", {})

is_gcs_bucket {
	input.subType == "gcp-gcs"
}

is_kms_key {
	input.subType == "gcp-kms"
}

is_bq_dataset {
	input.subType == "gcp-bq-dataset"
}

is_bq_table {
	input.subType == "gcp-bq-table"
}
