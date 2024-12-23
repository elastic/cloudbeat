package compliance.policy.aws_ecr.ensure_image_scan

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_ecr.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_aws_ecr

	rule_evaluation := data_adapter.image_scan_config.ScanOnPush

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"repository_name:": data_adapter.repository_name,
			"image_scanning_configuration": data_adapter.image_scan_config,
		},
	)
}
