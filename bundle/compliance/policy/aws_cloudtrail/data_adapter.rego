package compliance.policy.aws_cloudtrail.data_adapter

is_trail_type {
	input.subType = "aws-trail"
}

is_single_trail {
	input.subType == "aws-trail"
	not trail_items
}

trail = input.resource.Trail

trail_status = input.resource.Status

trail_bucket_info = input.resource.bucket_info

event_selectors = input.resource.EventSelectors

trail_items = input.resource.Items
