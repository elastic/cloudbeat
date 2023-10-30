package compliance.cis_gcp.rules.cis_4_2

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.compute.ensure_default_sa as audit
import data.compliance.policy.gcp.data_adapter

# Ensure That Instances Are Not Configured To Use the Default Service Account With Full Access to All Cloud APIs.
finding = result {
	# filter
	data_adapter.is_compute_instance

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(audit.sa_is_default_with_full_access)),
		{"Compute instance": input.resource},
	)
}
