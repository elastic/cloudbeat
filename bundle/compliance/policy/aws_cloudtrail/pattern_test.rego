package compliance.policy.aws_cloudtrail.pattern

filter_1 = {"FilterPattern": "filter_1"}

filter_2 = {"FilterPattern": "filter_2"}

pattern_1 = "filter_1"

pattern_2 = "filter_2"

pattern_never_match = "not_match"

test_pass {
	at_least_one_metric_exists({"MetricFilters": [filter_1]}, [pattern_1])
	at_least_one_metric_exists({"MetricFilters": [filter_1, filter_2]}, [pattern_1])
	at_least_one_metric_exists({"MetricFilters": [filter_1, filter_2]}, [pattern_2])
	at_least_one_metric_exists({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match, pattern_1])
}

test_fail {
	not at_least_one_metric_exists({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match])
	not at_least_one_metric_exists({"MetricFilters": []}, [])
}
