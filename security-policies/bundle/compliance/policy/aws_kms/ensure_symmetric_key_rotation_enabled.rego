package compliance.policy.aws_kms.ensure_symmetric_key_rotation_enabled

import data.compliance.lib.common as common
import data.compliance.policy.aws_kms.data_adapter

default rule_evaluation = false

rule_evaluation {
	data_adapter.key_rotation_enabled == true
}

finding = result {
	data_adapter.is_kms

	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{
			"KeyMetadata": data_adapter.key_metadata,
			"KeyRotationEnabled": data_adapter.key_rotation_enabled,
		},
	)
}
