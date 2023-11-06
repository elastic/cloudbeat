package compliance.cis_gcp.rules.cis_4_11

import data.compliance.lib.common
import data.compliance.policy.gcp.common as gcp_common
import data.compliance.policy.gcp.data_adapter

default is_confidential_computing_enabled = false

# Ensure That Compute Instances Have Confidential Computing Enabled.
finding = result {
	# filter
	data_adapter.is_compute_instance

	# confidential Computing is currently only supported on N2D machines
	startswith(gcp_common.get_machine_type_family(data_adapter.resource.data.machineType), "n2d-")

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_confidential_computing_enabled),
		{"Compute instance": input.resource},
	)
}

is_confidential_computing_enabled := data_adapter.resource.data.confidentialInstanceConfig.enableConfidentialCompute
