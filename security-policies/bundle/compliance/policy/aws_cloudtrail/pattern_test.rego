package compliance.policy.aws_cloudtrail.pattern

filter_1 = {"FilterPattern": "filter_1", "FilterName": "filter_1"}

filter_2 = {"FilterPattern": "filter_2", "FilterName": "filter_2"}

pattern_1 = "filter_1"

pattern_2 = "filter_2"

pattern_never_match = "not_match"

test_pass {
	get_filter_matched_to_pattern({"MetricFilters": [filter_1]}, [pattern_1])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_1])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_2])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match, pattern_1])
}

test_fail {
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match]) == ""
	get_filter_matched_to_pattern({"MetricFilters": []}, []) == ""
}
