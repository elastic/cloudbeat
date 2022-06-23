package compliance.cis_eks.rules.cis_5_1_1

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

# Check if image ScanOnPush is enabled
finding = result {
	# filter
	data_adapter.is_aws_ecr

	rule_evaluation := input.resource.ImageScanningConfiguration.ScanOnPush

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"repository_name:": input.resource.RepositoryName,
			"image_scanning_configuration": input.resource.ImageScanningConfiguration,
		},
	}
}
