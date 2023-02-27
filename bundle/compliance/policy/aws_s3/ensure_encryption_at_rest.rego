package compliance.policy.aws_s3.ensure_encryption_at_rest

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter

default rule_evaluation = false

rule_evaluation {
	data_adapter.sse_algorithm == "AES256"
}

rule_evaluation {
	data_adapter.sse_algorithm == "aws:kms"
}

finding = result {
	data_adapter.is_s3
	not data_adapter.sse_algorithm == null

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"SSEAlgorithm": data_adapter.sse_algorithm},
	)
}
