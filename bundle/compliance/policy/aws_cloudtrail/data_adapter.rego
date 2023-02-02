package compliance.policy.aws_cloudtrail.data_adapter

is_trail_type {
	input.subType = "aws-trail"
}

trail_items = input.resource.Items
