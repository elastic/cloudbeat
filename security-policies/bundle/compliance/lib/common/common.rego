package compliance.lib.common

import future.keywords.if
import future.keywords.in

# get OPA version
opa_version := opa.runtime().version

metadata := {
	"opa_version": opa_version,
	"policy_version": "1.0.0",
}

current_date := create_date_from_ns(time.now_ns())

past_date := "2021-12-25T12:43:00+00:00"

create_date_from_ns(x) := time_str if {
	date := time.date(x)
	t := time.clock(x)

	time_str := sprintf("%d-%02d-%02dT%02d:%02d:%02d+00:00", array.concat(date, t))
}

ConvertDaysToHours(duration) := result if {
	suffix := "d"
	contains(duration, suffix)
	days := trim_suffix(duration, suffix)
	result = sprintf("%dh", [to_number(days) * 24])
} else := duration

# set the rule result
calculate_result(evaluation) := "passed" if {
	evaluation != false
} else := "failed"

# Safely evaluate evidence. In case a key is undefined, it will be defaulted.
# keypaths is an object defined as {str: array}
collect_evidence(resource, key_paths) := {key: evidence |
	some key, path in key_paths
	evidence := object.get(resource, path, resource)
}

# get different evidence based on the evaluation
get_evidence(ok, passed, failed) := result if {
	ok == true
	result = passed
} else := failed

# If value is not an array, enclose it in one
ensure_array(value) := [value] if {
	not is_array(value)
} else := value

contains_key(object, key) if {
	object[key]
} else := false

contains_key_with_value(object, key, value) if {
	object[key] = value
} else := false

# checks if a value is greater or equals to a minimum value
greater_or_equal(value, minimum) if {
	to_number(value) >= minimum
} else := false

# checks if duration is less than some maximum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_lt(duration, max_duration) if {
	duration_ns := time.parse_duration_ns(duration)
	max_duration_ns := time.parse_duration_ns(max_duration)
	duration_ns < max_duration_ns
} else := false

# checks if duration is less than some maximum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_lte(duration, max_duration) if {
	duration_ns := time.parse_duration_ns(duration)
	max_duration_ns := time.parse_duration_ns(max_duration)
	duration_ns <= max_duration_ns
} else := false

# checks if duration is greater than some minimum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_gt(duration, min_duration) if {
	duration_ns := time.parse_duration_ns(duration)
	min_duration_ns := time.parse_duration_ns(min_duration)
	duration_ns > min_duration_ns
} else := false

# checks if duration is greater or equal to some minimum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_gte(duration, min_duration) if {
	duration_ns := time.parse_duration_ns(duration)
	min_duration_ns := time.parse_duration_ns(min_duration)
	duration_ns >= min_duration_ns
} else := false

# The function determines whether the given date occurs within the provided time period.
# date: time in nanoseconds
date_within_duration(date, duration) if {
	now = time.now_ns()
	duration_ns := time.parse_duration_ns(duration)
	date > now - duration_ns
} else := false

ranges_smaller_than(ranges, value) if {
	range := ranges[_]
	range < value
}

ranges_gte(ranges, value) if {
	not ranges_smaller_than(ranges, value)
}

generate_result(evaluation, evidence, expected) := {
	"evaluation": evaluation,
	"evidence": evidence,
	"expected": expected,
}

generate_result_without_expected(evaluation, evidence) := {
	"evaluation": evaluation,
	"evidence": evidence,
}
