package compliance.cis_azure.rules.cis_2_1_19

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	# filter
	data_adapter.is_security_contacts

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(owner_enabled),
		{"Resource": data_adapter.resource},
	)
}

default owner_enabled := false

owner_enabled if {
	# Ensure at least one Security Contact Settings exists and owner is selected.
	some security_contact in data_adapter.resource

	security_contact.name == "default"
	is_string(security_contact.properties.emails)
	security_contact.properties.emails != ""
}
