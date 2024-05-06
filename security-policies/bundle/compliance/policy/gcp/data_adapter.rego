package compliance.policy.gcp.data_adapter

import future.keywords.if
import future.keywords.in

resource := input.resource.resource

iam_policy := input.resource.iam_policy

has_policy := "iam_policy" in object.keys(input.resource)

is_policies_resource if {
	input.subType == "gcp-policies"
}

is_gke_instance(instance) if {
	startswith(instance.name, "gke-")
}

is_api_key if {
	input.subType == "gcp-apikeys-key"
}

is_dataproc_cluster if {
	input.subType == "gcp-dataproc-cluster"
}

is_storage_bucket if {
	input.subType == "gcp-storage-bucket"
}

is_cloud_resource_manager_project if {
	input.subType == "gcp-cloudresourcemanager-project"
}

is_iam_service_account if {
	input.subType == "gcp-iam-service-account"
}

is_iam_service_account_key if {
	input.subType == "gcp-iam-service-account-key"
}

is_cloudkms_crypto_key if {
	input.subType == "gcp-cloudkms-crypto-key"
}

is_bigquery_dataset if {
	input.subType == "gcp-bigquery-dataset"
}

is_bigquery_table if {
	input.subType == "gcp-bigquery-table"
}

is_compute_instance if {
	input.subType == "gcp-compute-instance"
}

is_firewall_rule if {
	input.subType == "gcp-compute-firewall"
}

is_compute_disk if {
	input.subType == "gcp-compute-disk"
}

is_compute_network if {
	input.subType == "gcp-compute-network"
}

is_cloud_sql if {
	input.subType == "gcp-sqladmin-instance"
}

is_sql_server if {
	startswith(resource.data.databaseVersion, "SQLSERVER")
}

is_cloud_my_sql if {
	startswith(resource.data.databaseVersion, "MYSQL")
}

is_backend_service if {
	input.subType == "gcp-compute-region-backend-service"
}

is_https_lb if {
	resource.data.protocol == "HTTPS"
}

is_postgres_sql if {
	startswith(resource.data.databaseVersion, "POSTGRES")
}

is_monitoring_asset if {
	input.subType == "gcp-monitoring"
}

is_dns_managed_zone if {
	input.subType == "gcp-dns-managed-zone"
}

is_sql_instance if {
	input.subType == "gcp-sqladmin-instance"
}

is_subnetwork if {
	input.subType == "gcp-compute-subnetwork"
}

is_log_bucket if {
	input.subType == "gcp-logging-log-bucket"
}

is_services_usage if {
	input.subType == "gcp-service-usage"
}

is_logging_asset if {
	input.subType == "gcp-logging"
}
