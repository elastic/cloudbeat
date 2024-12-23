package compliance.cis_gcp.rules.cis_3_1

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure That the Default Network Does Not Exist in a Project.
finding := result if {
	# filter
	data_adapter.is_compute_network

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_not_default_network),
		data_adapter.resource,
	)
}

is_not_default_network if {
	not data_adapter.resource.data.name == "default"
} else := false
