package compliance.lib.common

import data.compliance.lib.assert

test_calculate_result_rule_evaluation_false {
	rule_evaluation := false
	calculate_result(rule_evaluation) == "failed"
}

test_calculate_result_rule_evaluation_true {
	rule_evaluation := true
	calculate_result(rule_evaluation) == "passed"
}

test_array_contains {
	array := ["a", "b", "c"]
	key := "c"
	array_contains(array, key)
}

test_array_contains_not_contains {
	array := ["a", "b", "c"]
	key := "d"
	assert.is_false(array_contains(array, key))
}

test_contains_key {
	array := {"a": "aa", "b": "bb"}
	key := "a"
	contains_key(array, key)
}

test_contains_key_not_contains {
	array := {"a": "aa", "b": "bb"}
	key := "c"
	assert.is_false(contains_key(array, key))
}

test_greater_or_equal_greater {
	value := 10
	minimum := 9
	greater_or_equal(value, minimum)
}

test_greater_or_equal_equal {
	value := 10
	minimum := 10
	greater_or_equal(value, minimum)
}

test_greater_or_equal_smaller {
	value := 10
	minimum := 11
	assert.is_false(greater_or_equal(value, minimum))
}

test_duration_gt_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gt(duration, min_duration)
}

test_duration_gt_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_gt_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_lt_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lt(duration, min_duration)
}

test_duration_gte_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gte(duration, min_duration))
}

test_duration_lte_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lte(duration, min_duration))
}

test_duration_lte_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}

test_duration_lte_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}
