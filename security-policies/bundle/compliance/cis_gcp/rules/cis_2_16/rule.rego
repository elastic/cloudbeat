package compliance.cis_gcp.rules.cis_2_16

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding = result if {
	data_adapter.is_backend_service
	data_adapter.is_https_lb

	result := common.generate_result_without_expected(
		common.calculate_result(is_logging_enabled),
		data_adapter.resource,
	)
}

is_logging_enabled if {
	data_adapter.resource.data.logConfig.enable
} else = false
