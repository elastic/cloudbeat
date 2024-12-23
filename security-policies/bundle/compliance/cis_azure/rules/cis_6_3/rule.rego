package compliance.cis_azure.rules.cis_6_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.virtual_machine.network_rules as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vm

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(udp_ports_closed),
		{"Resource": data_adapter.resource},
	)
}

default udp_ports_closed := false

udp_ports_closed if {
	audit.vm_has_closed_port(data_adapter, "53", "UDP")
	audit.vm_has_closed_port(data_adapter, "123", "UDP")
	audit.vm_has_closed_port(data_adapter, "161", "UDP")
	audit.vm_has_closed_port(data_adapter, "389", "UDP")
	audit.vm_has_closed_port(data_adapter, "1900", "UDP")
}
