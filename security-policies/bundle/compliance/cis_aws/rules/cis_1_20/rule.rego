package compliance.cis_aws.rules.cis_1_20

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.every
import future.keywords.if
import future.keywords.in

# Ensure that IAM Access analyzer is enabled for all regions
finding := result if {
	# filter
	data_adapter.is_access_analyzers

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(analyzer_exists),
		{"Access Analyzers": input.resource},
	)
}

analyzer_exists if {
	every region in data_adapter.analyzer_regions {
		some analyzer in data_adapter.analyzers
		analyzer.Region == region
		analyzer.Status == "ACTIVE"
	}
} else := false
