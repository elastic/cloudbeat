package compliance.cis_azure.rules.cis_6_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.virtual_machine.network_rules as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vm

	portProperConfigured := audit.vm_has_closed_port(data_adapter, "80", "TCP")

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(portProperConfigured),
		{"Resource": data_adapter.resource},
	)
}
