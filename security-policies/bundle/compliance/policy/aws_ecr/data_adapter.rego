package compliance.policy.aws_ecr.data_adapter

import future.keywords.if

is_aws_ecr if {
	input.subType == "aws-ecr"
}

cluster := input.resource.Cluster

image_scan_config := input.resource.ImageScanningConfiguration

repository_name := input.resource.RepositoryName
