package compliance.policy.aws_cloudtrail.data_adapter

is_multi_trails_type {
	input.subType = "aws-multi-trails"
}

is_single_trail {
	input.subType == "aws-trail"
}

trail = input.resource.Trail

trail_status = input.resource.Status

trail_bucket_info = input.resource.bucket_info

event_selectors = input.resource.EventSelectors

trail_items = input.resource.Items
