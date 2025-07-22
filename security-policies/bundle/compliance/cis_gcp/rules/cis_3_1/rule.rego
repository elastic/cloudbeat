package compliance.cis_gcp.rules.cis_3_1

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure That the Default Network Does Not Exist in a Project.
finding := result if {
	# filter
	data_adapter.is_compute_network

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_not_default_network))
}

is_not_default_network if {
	not data_adapter.resource.data.name == "default"
} else := false
