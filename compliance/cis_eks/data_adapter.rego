package compliance.cis_eks.data_adatper

is_aws_eks {
	input.type == "aws-eks"
}

is_aws_elb {
	input.type == "aws-elb"
}

is_aws_ecr {
	input.type == "aws-ecr"
}
