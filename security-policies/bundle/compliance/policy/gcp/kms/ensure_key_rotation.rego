package compliance.policy.gcp.kms.ensure_key_rotation

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

duration := sprintf("%dh", [90 * 24]) # 90 days converted to hours

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_cloudkms_crypto_key

	# In order for an encryption key to be available,
	# it needs to have a primary key version which is enabled
	data_adapter.resource.data.primary.state == "ENABLED"

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"KMS": data_adapter.resource},
	)
}

rule_evaluation if {
	common.date_within_duration(time.parse_rfc3339_ns(data_adapter.resource.data.nextRotationTime), duration)
	common.duration_lte(common.ConvertDaysToHours(data_adapter.resource.data.rotationPeriod), duration)
}
