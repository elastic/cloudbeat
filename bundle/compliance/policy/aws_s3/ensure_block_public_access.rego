package compliance.policy.aws_s3.ensure_block_public_access

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter
import future.keywords.in

default rule_evaluation = false

rule_evaluation {
	data_adapter.public_access_block_configuration.BlockPublicAcls == true
	data_adapter.public_access_block_configuration.BlockPublicPolicy == true
	data_adapter.public_access_block_configuration.IgnorePublicAcls == true
	data_adapter.public_access_block_configuration.RestrictPublicBuckets == true
}

finding = result {
	data_adapter.is_s3
	not data_adapter.public_access_block_configuration == null

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"PublicAccessBlockConfiguration": data_adapter.public_access_block_configuration},
	)
}
