package compliance.cis_azure.rules.cis_5_1_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	# filter
	data_adapter.is_diagnostic_settings

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(required_categories_enabled),
		{"Resource": data_adapter.resource},
	)
}

default required_categories_enabled := false

required_categories_enabled if {
	diagnostic_settings_category_enabled("Administrative") == true
	diagnostic_settings_category_enabled("Alert") == true
	diagnostic_settings_category_enabled("Policy") == true
	diagnostic_settings_category_enabled("Security") == true
}

diagnostic_settings_category_enabled(category) if {
	# Ensure at least one Diagnostic Settings entry passes the following conditions
	some diagnostic_settings in data_adapter.diagnostic_settings

	# Ensure the category is enable in at least one diagnostic settings
	category_is_enabled := [log |
		log := diagnostic_settings.properties.logs[_]
		log.category == category
		log.enabled == true
	]
	count(category_is_enabled) >= 1
} else := false
