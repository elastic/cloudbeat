package compliance.cis_gcp.rules.cis_4_9

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.in

# Ensure That Compute Instances Do Not Have Public IP Addresses.
finding = result {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(is_publicly_exposed)),
		{"Compute instance": input.resource},
	)
}

is_publicly_exposed {
	some networkInterface in data_adapter.resource.data.networkInterfaces
	networkInterface.accessConfigs
} else = false
