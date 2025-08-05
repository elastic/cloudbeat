package compliance.cis_gcp.rules.cis_3_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure Legacy Networks Do Not Exist for Older Projects
finding := result if {
	data_adapter.is_compute_network

	result := common.generate_evaluation_result(common.calculate_result(is_not_legacy_network))
}

is_not_legacy_network if {
	# When autoCreateSubnetworks is set to false a legacy network is being created (https://cloud.google.com/compute/docs/reference/rest/v1/networks).
	data_adapter.resource.data.autoCreateSubnetworks
} else := false
