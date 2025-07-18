package compliance.cis_gcp.rules.cis_4_8

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure Compute Instances Are Launched With Shielded VM Enabled.
finding := result if {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_shielded_vm))
}

is_shielded_vm if {
	cfg := data_adapter.resource.data.shieldedInstanceConfig
	cfg.enableIntegrityMonitoring
	cfg.enableVtpm
} else := false
