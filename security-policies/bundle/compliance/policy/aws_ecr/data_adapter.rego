package compliance.policy.aws_ecr.data_adapter

is_aws_ecr {
	input.subType == "aws-ecr"
}

cluster = input.resource.Cluster

image_scan_config = input.resource.ImageScanningConfiguration

repository_name = input.resource.RepositoryName
