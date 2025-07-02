package compliance.cis_gcp.rules.cis_4_9

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure That Compute Instances Do Not Have Public IP Addresses.
finding := result if {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_evaluation_result(common.calculate_result(assert.is_false(is_publicly_exposed)))
}

is_publicly_exposed if {
	some networkInterface in data_adapter.resource.data.networkInterfaces
	networkInterface.accessConfigs
} else := false
