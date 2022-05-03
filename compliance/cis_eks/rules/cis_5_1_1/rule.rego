package compliance.cis_eks.rules.cis_5_1_1

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common

default rule_evaluation = false

# Checks that every repository scanOnPush is enabled
rule_evaluation {
	input.resource.EcrRepositories

	# Verify there is no unsafe image
	misconfigured_repositories = [index | assert.is_false(input.resource.EcrRepositories[index].ImageScanningConfiguration.ScanOnPush)]
	count(misconfigured_repositories) == 0
}

evidence["misconfigured_repositories"] = misconfigured_repo {
	misconfigured_repo = [repo |
		repo := input.resource.EcrRepositories[index].RepositoryName
		assert.is_false(input.resource.EcrRepositories[index].ImageScanningConfiguration.ScanOnPush)
	]
}

# Check if image ScanOnPush is enabled
finding = result {
	# filter
	data_adapter.is_aws_ecr

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": evidence,
	}
}
