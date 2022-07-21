package compliance.cis_eks.data_adapter

import data.compliance.lib.data_adapter

is_aws_eks {
	input.subType == "aws-eks"
}

is_aws_elb {
	input.subType == "aws-elb"
}

is_aws_ecr {
	input.subType == "aws-ecr"
}

process_args = result {
	result = data_adapter.process_args(" ")
}
