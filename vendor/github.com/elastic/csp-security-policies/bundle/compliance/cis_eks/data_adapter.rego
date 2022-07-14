package compliance.cis_eks.data_adapter

import data.compliance.lib.data_adapter

is_aws_eks {
	input.type == "aws-eks"
}

is_aws_elb {
	input.type == "aws-elb"
}

is_aws_ecr {
	input.type == "aws-ecr"
}

process_args = result {
	result = data_adapter.process_args(" ")
}
