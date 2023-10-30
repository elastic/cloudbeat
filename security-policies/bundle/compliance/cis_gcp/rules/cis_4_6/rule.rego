package compliance.cis_gcp.rules.cis_4_6

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

default is_ip_forwarding_enabled = false

# Ensure That IP Forwarding Is Not Enabled on Instances
finding = result {
	# filter
	data_adapter.is_compute_instance

	# VMs created by GKE should be excluded
	not data_adapter.is_gke_instance(data_adapter.resource.data)

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(is_ip_forwarding_enabled)),
		{"Compute instance": input.resource},
	)
}

is_ip_forwarding_enabled := data_adapter.resource.data.canIpForward
