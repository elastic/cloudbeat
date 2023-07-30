package compliance.cis_gcp.rules.cis_1_7

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default is_valid = false

duration = sprintf("%dh", [90 * 24]) # 90 days converted to hours

is_valid if {
	date := time.parse_rfc3339_ns(data_adapter.resource.data.validAfterTime)
	common.date_within_duration(date, duration)
}

finding = result if {
	data_adapter.is_iam_service_account_key

	result := common.generate_result_without_expected(
		common.calculate_result(is_valid),
		data_adapter.resource.data.validAfterTime,
	)
}
