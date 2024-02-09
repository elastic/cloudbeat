package compliance.policy.aws_cloudtrail.pattern

import future.keywords.if

filter_1 = {"FilterPattern": "a=b", "FilterName": "filter_1"}

filter_2 = {"FilterPattern": "b=c", "FilterName": "filter_2"}

pattern_1 = "a=b"

pattern_2 = "b=c"

pattern_never_match = "not_match"

test_pass if {
	get_filter_matched_to_pattern({"MetricFilters": [filter_1]}, [pattern_1])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_1])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_2])
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match, pattern_1])
}

test_fail if {
	get_filter_matched_to_pattern({"MetricFilters": [filter_1, filter_2]}, [pattern_never_match]) == ""
	get_filter_matched_to_pattern({"MetricFilters": []}, []) == ""
}

# regal ignore:rule-length
test_expressions_equivalent if {
	is_equal(
		"TRUE simple expression equal", expressions_equivalent(
			simple_expression("a", "=", "b"),
			simple_expression("a", "=", "b"),
		),
		true,
	)

	is_equal(
		"TRUE NOT EXISTS equal", expressions_equivalent(
			simple_expression("a", "NOT EXISTS", ""),
			simple_expression("a", "NOT EXISTS", ""),
		),
		true,
	)

	is_equal(
		"TRUE simple expression inverted", expressions_equivalent(
			simple_expression("a", "=", "b"),
			simple_expression("b", "=", "a"),
		),
		true,
	)

	is_equal(
		"FALSE simple expression different left", expressions_equivalent(
			simple_expression("a", "=", "b"),
			simple_expression("b", "=", "f"),
		),
		false,
	)

	is_equal(
		"FALSE simple expression different right", expressions_equivalent(
			simple_expression("a", "=", "b"),
			simple_expression("c", "=", "b"),
		),
		false,
	)

	is_equal(
		"FALSE simple expression different operator", expressions_equivalent(
			simple_expression("a", "=", "b"),
			simple_expression("b", "!=", "b"),
		),
		false,
	)

	is_equal(
		"TRUE complex expression one level basic equal", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
		),
		true,
	)

	is_equal(
		"TRUE complex expression one level basic inverted", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
			complex_expression("||", [
				simple_expression("d", "=", "e"),
				simple_expression("a", "=", "b"),
			]),
		),
		true,
	)

	is_equal(
		"TRUE complex expression one level basic inverted (also sub expressions)", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
			complex_expression("||", [
				simple_expression("d", "=", "e"),
				simple_expression("b", "=", "a"),
			]),
		),
		true,
	)

	is_equal(
		"FALSE complex expression one level global different op", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
			complex_expression("&&", [
				simple_expression("a", "=", "b"),
				simple_expression("d", "=", "e"),
			]),
		),
		false,
	)

	is_equal(
		"TRUE complex expression two levels equal", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [
					simple_expression("d", "=", "e"),
					simple_expression("f", "=", "g"),
				]),
			]),
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [
					simple_expression("d", "=", "e"),
					simple_expression("f", "=", "g"),
				]),
			]),
		),
		true,
	)

	is_equal(
		"TRUE complex expression two levels different orders", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [
					simple_expression("d", "=", "e"),
					simple_expression("f", "=", "g"),
				]),
			]),
			complex_expression("||", [
				complex_expression("&&", [
					simple_expression("g", "=", "f"),
					simple_expression("d", "=", "e"),
				]),
				simple_expression("b", "=", "a"),
			]),
		),
		true,
	)

	is_equal(
		"FALSE complex expression two levels different comparisson", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [
					simple_expression("d", "=", "DIFFERENT"),
					simple_expression("f", "=", "g"),
				]),
			]),
			complex_expression("||", [
				complex_expression("&&", [
					simple_expression("g", "=", "f"),
					simple_expression("d", "=", "e"),
				]),
				simple_expression("b", "=", "a"),
			]),
		),
		false,
	)

	is_equal(
		"FALSE complex expression two levels missing one sub expression", expressions_equivalent(
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [
					simple_expression("d", "=", "e"),
					simple_expression("f", "=", "g"),
				]),
			]),
			complex_expression("||", [
				simple_expression("a", "=", "b"),
				complex_expression("&&", [simple_expression("f", "=", "g")]),
			]),
		),
		false,
	)
}

is_equal(_, actual, want) if {
	actual == want
}

is_equal(desc, actual, want) if {
	actual != want

	print("--- Test [", desc, "] failed because:") # regal ignore:print-or-trace-call
	print("WANT:    ", want) # regal ignore:print-or-trace-call
	print("ACTUAL:  ", actual) # regal ignore:print-or-trace-call

	# Force failure
	actual == want
}
