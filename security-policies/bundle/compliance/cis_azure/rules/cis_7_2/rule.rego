package compliance.cis_azure.rules.cis_7_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vm

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(has_managed_disk),
		{"Resource": data_adapter.resource},
	)
}

has_managed_disk if {
	data_adapter.properties.storageProfile.osDisk.managedDisk.id != ""
} else := false
