package compliance.cis_azure.rules.cis_1_23

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every

finding = result {
	# filter
	data_adapter.is_role_definition

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(no_custom_roles),
		{"Resource": data_adapter.resource},
	)
}

no_custom_roles {
	every role_def in data_adapter.role_definitions {
		role_def.properties.type != "CustomRole"
	}
} else = false
