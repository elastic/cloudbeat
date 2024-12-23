package compliance.cis_azure.rules.cis_5_3_1

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	# filter
	data_adapter.is_insights_component

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(component_exists),
		{"Resource": data_adapter.resource},
	)
}

component_exists if {
	some insights_component in data_adapter.insights_components
	insights_component.properties.provisioningState == "Succeeded"
} else := false
