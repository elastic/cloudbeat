package compliance.policy.aws_cloudtrail.ensure_cloudwatch

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.if

default ensure_cloudwatch_logs_enabled := false

ensure_cloudwatch_logs_enabled if {
	data_adapter.trail.CloudWatchLogsLogGroupArn != ""
	common.date_within_duration(time.parse_rfc3339_ns(data_adapter.trail_status.LatestCloudWatchLogsDeliveryTime), "24h")
}
