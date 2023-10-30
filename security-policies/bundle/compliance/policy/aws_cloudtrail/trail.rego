package compliance.policy.aws_cloudtrail.trail

import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern

# Satisfied trail is one that
# 1. is multi region trail
# 2. the logging is enabled
# 3. captures all management events
# 4. has metric filter as expected
# 5. has sns topic subscription
at_least_one_trail_satisfied(metric_filter_patterns) {
	some i
	trail := data_adapter.trail_items[i]
	cloudtrail_enabled(trail)

	# and the metric filter pattern is as expected
	filter := pattern.get_filter_matched_to_pattern(trail, metric_filter_patterns)

	# and it has at least one subscription
	count(trail.MetricTopicBinding[filter]) > 0
}

cloudtrail_enabled(trail) {
	# is multi region trail
	trail.TrailInfo.Trail.IsMultiRegionTrail

	# and it is avtive
	trail.TrailInfo.Status.IsLogging

	# and it captures all management events
	some i
	selector := trail.TrailInfo.EventSelectors[i]
	selector.IncludeManagementEvents
	selector.ReadWriteType == "All"
}
