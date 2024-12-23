package compliance.cis_gcp.rules.cis_4_3

import data.compliance.lib.common
import data.compliance.policy.gcp.compute.assess_instance_metadata as audit
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure “Block Project-Wide SSH Keys” Is Enabled for VM Instances
finding := result if {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_project_ssh_keys_enabled),
		{"Compute instance": input.resource},
	)
}

is_project_ssh_keys_enabled := audit.is_instance_metadata_valid("block-project-ssh-keys", "true")
