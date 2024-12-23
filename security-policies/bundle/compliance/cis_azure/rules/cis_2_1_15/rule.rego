package compliance.cis_azure.rules.cis_2_1_15

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	# filter
	data_adapter.is_security_auto_provisioning_settings

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(auto_provisioning_on),
		{"Resource": data_adapter.resource},
	)
}

default auto_provisioning_on := false

auto_provisioning_on if {
	# Ensure at least one Auto Provisioning Settings exists and autoProvision is set to on.
	some auto_provisioning_settings in data_adapter.resource

	auto_provisioning_settings.name == "default"
	lower(auto_provisioning_settings.properties.autoProvision) == "on"
}
