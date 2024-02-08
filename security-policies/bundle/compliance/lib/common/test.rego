package compliance.lib.common

import data.compliance.lib.assert
import future.keywords.if

test_calculate_result_rule_evaluation_false if {
	rule_evaluation := false
	calculate_result(rule_evaluation) == "failed"
}

test_calculate_result_rule_evaluation_true if {
	rule_evaluation := true
	calculate_result(rule_evaluation) == "passed"
}

test_collect_evidence_keys if {
	resource := {
		"type": "escape-pod",
		"contents": "no life-forms aboard",
		"crew": ["R2-D2", "C-3PO"],
	}
	collect_evidence(resource, {"Type": "type", "Crew": "crew"}) == {
		"Type": resource.type,
		"Crew": resource.crew,
	}
}

test_collect_evidence_nested_keys if {
	resource := {"a": {"b": {"c": {"d": "nested_value"}}}}
	collect_evidence(resource, {"D": ["a", "b", "c", "d"]}) == {"D": resource.a.b.c.d}
}

test_collect_evidence_array_element if {
	resource := {
		"type": "escape-pod",
		"contents": "no life-forms aboard",
		"crew": ["R2-D2", "C-3PO"],
	}
	collect_evidence(resource, {"Second Robot": ["crew", 1]}) == {"Second Robot": "C-3PO"}
}

test_collect_evidence_empty_keypaths if {
	collect_evidence({}, {}) == {}
}

test_collect_evidence_self_referential_keypaths if {
	resource := {"a": 1, "b": 2}
	collect_evidence(resource, {"self": []}) == {"self": resource}
}

test_collect_evidence_blank_paths if {
	resource := {"a": 1, "b": 2}
	collect_evidence(resource, {"empty": ""}) == {"empty": resource}
}

test_collect_evidence_non_existent_key if {
	resource := {"a": 1, "b": 2}
	collect_evidence(resource, {"C": "c"}) == {"C": resource}
}

test_collect_evidence_non_existent_path_tail if {
	resource := {"a": {"b": {"c": 42}}}
	collect_evidence(resource, {"a.b.x": ["a", "b", "x"]}) == {"a.b.x": resource}
}

test_collect_evidence_non_existent_path_segment if {
	resource := {"a": {"b": {"c": 42}}}
	collect_evidence(resource, {"a.x.c": ["a", "x", "c"]}) == {"a.x.c": resource}
}

test_collect_evidence_with_out_of_bounds_array_index if {
	resource := {"arr": [0, 1, 2, 3, 4]}
	collect_evidence(resource, {"arr[99]": ["arr", 99]}) == {"arr[99]": resource}
}

test_ensure_array_empty if {
	ensure_array([]) == []
}

test_ensure_array_from_array if {
	array := ["a", "b", "c"]
	ensure_array(array) == array
}

test_ensure_array_from_int if {
	ensure_array(1) == [1]
}

test_ensure_array_from_null if {
	ensure_array(null) == [null]
}

test_ensure_array_from_string if {
	ensure_array("a") == ["a"]
}

test_greater_or_equal_greater if {
	value := 10
	minimum := 9
	greater_or_equal(value, minimum)
}

test_greater_or_equal_equal if {
	value := 10
	minimum := 10
	greater_or_equal(value, minimum)
}

test_greater_or_equal_smaller if {
	value := 10
	minimum := 11
	assert.is_false(greater_or_equal(value, minimum))
}

test_duration_gt_greater if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gt(duration, min_duration)
}

test_duration_gt_equals if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_gt_smaller if {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_lt_greater if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_equals if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_smaller if {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lt(duration, min_duration)
}

test_duration_gte_greater if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_equals if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_smaller if {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gte(duration, min_duration))
}

test_duration_lte_greater if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lte(duration, min_duration))
}

test_duration_lte_equals if {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}

test_duration_lte_smaller if {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}

test_date_within_duration_later_than_threshold if {
	date := time.add_date(time.now_ns(), 0, 0, -1) # years, months, days
	date_within_duration(date, "48h")
}

test_date_within_duration_earlier_than_threshold if {
	date := time.add_date(time.now_ns(), 0, 0, -3) # years, months, days
	assert.is_false(date_within_duration(date, "48h"))
}
