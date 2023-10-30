package compliance.cis_azure.rules.cis_7_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter

finding = result {
	# filter
	data_adapter.is_bastion

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(at_least_one_bastion),
		{"Resource": data_adapter.resource},
	)
}

at_least_one_bastion {
	some i
	data_adapter.bastions[i].id != ""
} else = false
