package compliance.policy.aws_cloudtrail.pattern

# get a filter from a trail has at least one metric filter pattern that matches at least one pattern
get_filter_matched_to_pattern(trail, patterns) = name {
	some i, j
	filter := trail.MetricFilters[i]
	pattern := patterns[j]
	filter.FilterPattern == pattern
	name := filter.FilterName
} else = ""
