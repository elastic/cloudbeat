package compliance.cis_eks.data_adapter

is_aws_elb {
	input.subType == "aws-elb"
}

is_aws_ecr {
	input.subType == "aws-ecr"
}

process_args_seperator = " "
