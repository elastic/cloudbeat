package compliance.policy.aws_kms.ensure_symmetric_key_rotation_enabled

import data.compliance.lib.common
import data.compliance.policy.aws_kms.data_adapter
import future.keywords.if

default rule_evaluation := false

rule_evaluation if {
	data_adapter.key_rotation_enabled == true
}

finding := result if {
	data_adapter.is_kms

	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{
			"KeyMetadata": data_adapter.key_metadata,
			"KeyRotationEnabled": data_adapter.key_rotation_enabled,
		},
	)
}
