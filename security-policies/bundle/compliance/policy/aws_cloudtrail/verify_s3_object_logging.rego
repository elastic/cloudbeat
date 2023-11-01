package compliance.policy.aws_cloudtrail.verify_s3_object_logging

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter

#1.Checks if trail is multi region
#2.Checks if trail has an event selector of "allowed_types" type.
#3.Checks if the type of data resource is "AWS::S3::Object" (S3 object).
#4.Checks if the partial ARN of the data resource is "arn:aws:s3".
ensure_s3_object_logging(allowed_types) {
	data_adapter.trail.IsMultiRegionTrail

	some i, j, k
	selector := data_adapter.event_selectors[i]
	common.array_contains(allowed_types, selector.ReadWriteType)

	dataResource := selector.DataResources[j]
	dataResource.Type == "AWS::S3::Object"

	partialARN := dataResource.Values[k]
	startswith(partialARN, "arn:aws:s3")
} else = false
