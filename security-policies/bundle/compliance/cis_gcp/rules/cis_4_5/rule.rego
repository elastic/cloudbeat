package compliance.cis_gcp.rules.cis_4_5

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.compute.assess_instance_metadata as audit
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure ‘Enable Connecting to Serial Ports’ Is Not Enabled for VM Instance
finding := result if {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(is_serial_port_enabled)),
		{"Compute instance": input.resource},
	)
}

is_serial_port_enabled := audit.is_instance_metadata_valid("serial-port-enable", "true")
