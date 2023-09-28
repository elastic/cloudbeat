package compliance.policy.gcp.data_adapter

import data.compliance.lib.common

resource = input.resource.resource

iam_policy = input.resource.iam_policy

has_policy = common.contains_key(input.resource, "iam_policy")

is_policies_resource {
	input.subType == "gcp-policies"
}

is_gke_instance(instance) {
	startswith(instance.name, "gke-")
}

is_api_key {
	input.subType == "gcp-apikeys-key"
}

is_dataproc_cluster {
	input.subType == "gcp-dataproc-cluster"
}

is_storage_bucket {
	input.subType == "gcp-storage-bucket"
}

is_cloud_resource_manager_project {
	input.subType == "gcp-cloudresourcemanager-project"
}

is_iam_service_account {
	input.subType == "gcp-iam-service-account"
}

is_api_key {
	input.subType == "gcp-apikeys-key"
}

is_iam_service_account_key {
	input.subType == "gcp-iam-service-account-key"
}

is_cloudkms_crypto_key {
	input.subType == "gcp-cloudkms-crypto-key"
}

is_bigquery_dataset {
	input.subType == "gcp-bigquery-dataset"
}

is_bigquery_table {
	input.subType == "gcp-bigquery-table"
}

is_compute_instance {
	input.subType == "gcp-compute-instance"
}

is_firewall_rule {
	input.subType == "gcp-compute-firewall"
}

is_compute_disk {
	input.subType == "gcp-compute-disk"
}

is_compute_network {
	input.subType == "gcp-compute-network"
}

is_cloud_sql {
	input.subType == "gcp-sqladmin-instance"
}

is_sql_server {
	startswith(resource.data.databaseVersion, "SQLSERVER")
}

is_cloud_my_sql {
	startswith(resource.data.databaseVersion, "MYSQL")
}

is_backend_service {
	input.subType == "gcp-compute-region-backend-service"
}

is_https_lb {
	resource.data.protocol == "HTTPS"
}

is_postgres_sql {
	startswith(resource.data.databaseVersion, "POSTGRES")
}

is_monitoring_asset {
	input.subType == "gcp-monitoring"
}

is_dns_managed_zone {
	input.subType == "gcp-dns-managed-zone"
}

is_sql_instance {
	input.subType == "gcp-sqladmin-instance"
}

is_subnetwork {
	input.subType == "gcp-compute-subnetwork"
}

is_log_bucket {
	input.subType == "gcp-logging-log-bucket"
}

is_services_usage {
	input.subType == "gcp-service-usage"
}

is_logging_asset {
	input.subType == "gcp-logging"
}
