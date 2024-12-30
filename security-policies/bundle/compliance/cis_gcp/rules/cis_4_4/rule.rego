package compliance.cis_gcp.rules.cis_4_4

import data.compliance.lib.common
import data.compliance.policy.gcp.compute.assess_instance_metadata as audit
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure Oslogin Is Enabled for a Project
finding := result if {
	# filter
	data_adapter.is_compute_instance

	# VMs created by GKE should be excluded
	not data_adapter.is_gke_instance(data_adapter.resource.data)

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_oslogin_enabled),
		{"Compute instance": input.resource},
	)
}

is_oslogin_enabled := audit.is_instance_metadata_valid("enable-oslogin", "true")
