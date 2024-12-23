package compliance.cis_gcp.rules.cis_2_12

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure That Cloud DNS Logging Is Enabled for All VPC Networks.
finding := result if {
	data_adapter.is_compute_network

	result := common.generate_result_without_expected(
		common.calculate_result(is_dns_logging_enabled),
		data_adapter.resource,
	)
}

is_dns_logging_enabled if {
	data_adapter.resource.data.enabledDnsLogging == true
} else := false
