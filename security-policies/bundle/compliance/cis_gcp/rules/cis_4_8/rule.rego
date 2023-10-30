package compliance.cis_gcp.rules.cis_4_8

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

# Ensure Compute Instances Are Launched With Shielded VM Enabled.
finding = result {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_shielded_vm),
		{"Compute instance": input.resource},
	)
}

is_shielded_vm {
	cfg := data_adapter.resource.data.shieldedInstanceConfig
	cfg.enableIntegrityMonitoring
	cfg.enableVtpm
} else = false
