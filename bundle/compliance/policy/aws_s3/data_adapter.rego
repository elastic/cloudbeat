package compliance.policy.aws_s3.data_adapter

is_s3 {
	input.subType == "aws-s3"
}

sse_algorithm := input.resource.SSEAlgorithm
