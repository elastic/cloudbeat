package compliance.policy.aws_cloudtrail.pattern

import future.keywords.every
import future.keywords.if
import future.keywords.in

# get a filter from a trail has at least one metric filter pattern that matches at least one pattern
get_filter_matched_to_pattern(trail, patterns) = name if {
	some i, j
	filter := trail.MetricFilters[i]
	pattern := patterns[j]
	filter.FilterPattern == pattern
	name := filter.FilterName
} else = ""

default parse_expression(s) = {"error": "Could not parse expression"}

parse_expression(s) = expression if { # handle simple expression
	# Filter
	is_simple_expression(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Parse
	expression = parse_simple_expression(clean_s)
}

parse_expression(s) = expression if { # handle complex one level deep AND only complex expression
	# Filter
	is_complex_expression(s)
	has_and_operator(s)
	not has_or_operator(s)

	# Clean and Parse
	expression = parse_one_level_one_operator_complex_expression(s, "&&")
}

parse_expression(s) = expression if { # handle complex one level deep OR only complex expression
	# Filter
	is_complex_expression(s)
	not has_and_operator(s)
	has_or_operator(s)

	# Clean and Parse
	expression = parse_one_level_one_operator_complex_expression(s, "||")
}

parse_expression(s) = expression if { # handle complex 2 operator (&& as main operator)
	# Filter
	is_complex_expression(s)
	has_and_operator(s)
	has_or_operator(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Filter once it's clean
	is_main_operator_and(clean_s)

	# Process
	expression = parse_two_operators_expression(clean_s, "&&")
}

parse_expression(s) = expression if { # handle complex 2 operator (|| as main operator)
	# Filter
	is_complex_expression(s)
	has_and_operator(s)
	has_or_operator(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Filter once it's clean
	is_main_operator_or(clean_s)

	# Process
	expression = parse_two_operators_expression(clean_s, "||")
}

parse_two_operators_expression(clean_s, main_op) = expression if {
	broken_s = split(clean_s, main_op)
	expressions = [
	mapped |
		original := broken_s[_]
		mapped := parse_second_level(original)
	]

	expression = complex_expression(main_op, expressions)
}

default is_main_operator_and(s) = false

is_main_operator_and(s) if { # if it has no outer parenthesis every sub expression must be either simple or valid or
	split(s, "")[0] != "("
	sub_expressions = split(s, "&&")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_or_operator(exp)
	}
}

is_main_operator_and(s) if { # if its has outer unwrap it and all sub expessions
	split(s, "")[0] == "("
	unwrapped = substring(s, 1, count(s) - 2)
	is_parenthesis_balanced(unwrapped)
	sub_expressions = split(unwrapped, "&&")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_or_operator(exp)
	}
}

is_main_operator_and(s) if { # try to process with everything is good already in the first place
	sub_expressions = split(s, "&&")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_or_operator(exp)
	}
}

is_sub_expression_simple_or_has_or_operator(s) if {
	is_parenthesis_balanced(s)
	has_or_operator(s)
}

is_sub_expression_simple_or_has_or_operator(s) if {
	is_simple_expression(s)
}

default is_main_operator_or(s) = false

is_main_operator_or(s) if { # if it has no outer parenthesis every sub expression must be either simple or valid or
	split(s, "")[0] != "("
	sub_expressions = split(s, "||")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_and_operator(exp)
	}
}

is_main_operator_or(s) if { # if its has outer unwrap it and all sub expessions
	split(s, "")[0] == "("
	unwrapped = substring(s, 1, count(s) - 2)
	is_parenthesis_balanced(unwrapped)
	sub_expressions = split(unwrapped, "||")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_and_operator(exp)
	}
}

is_main_operator_or(s) if { # try to process with everything is good already in the first place
	sub_expressions = split(s, "||")
	every exp in sub_expressions {
		is_sub_expression_simple_or_has_and_operator(exp)
	}
}

is_sub_expression_simple_or_has_and_operator(s) if {
	is_parenthesis_balanced(s)
	has_and_operator(s)
}

is_sub_expression_simple_or_has_and_operator(s) if {
	is_simple_expression(s)
}

parse_second_level(s) = expression if {
	# Filter
	is_simple_expression(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Parse
	expression = parse_simple_expression(clean_s)
}

parse_second_level(s) = expression if { # handle complex one level deep AND only complex expression
	# Filter
	is_complex_expression(s)
	has_and_operator(s)
	not has_or_operator(s)

	# Clean and Parse
	expression = parse_one_level_one_operator_complex_expression(s, "&&")
}

parse_second_level(s) = expression if { # handle complex one level deep OR only complex expression
	# Filter
	is_complex_expression(s)
	not has_and_operator(s)
	has_or_operator(s)

	# Clean and Parse
	expression = parse_one_level_one_operator_complex_expression(s, "||")
}

parse_second_level(s) = expression if {
	# Filter
	is_simple_expression(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Parse
	expression = parse_simple_expression(clean_s)
}

parse_one_level_one_operator_complex_expression(s, op) = expression if {
	# Clean
	clean_s = clean_full_expression(s)

	# Parse
	subStrings = split(clean_s, op)
	expressions = [
	mapped |
		original := subStrings[_]
		mapped := parse_simple_expression(original)
	]

	expression = complex_expression(op, expressions)
}

has_and_operator(s) if {
	contains(s, "&&")
}

has_or_operator(s) if {
	contains(s, "||")
}

is_simple_expression(s) if {
	not is_complex_expression(s)
}

is_complex_expression(s) if {
	has_and_operator(s)
}

is_complex_expression(s) if {
	has_or_operator(s)
}

parse_simple_expression(s) = expression if { # not equals
	op := "!="
	indexof(s, op) > 0
	parts := split(s, op)
	count(parts) == 2
	expression := simple_expression(clean_left(parts[0]), op, clean_right(parts[1]))
}

parse_simple_expression(s) = expression if { # equals
	indexof(s, "!=") == -1 # safety to not collide with different implementation
	op := "="
	indexof(s, op) > 0
	parts := split(s, op)
	count(parts) == 2
	expression := simple_expression(clean_left(parts[0]), op, clean_right(parts[1]))
}

parse_simple_expression(s) = expression if { # not exists
	op := "NOT EXISTS"
	indexof(s, op) > 0
	parts := split(s, op)
	count(parts) > 1
	expression := simple_expression(clean_left(parts[0]), op, "")
}

clean_full_expression(s) = clean if {
	clean = remove_spaces_between_parenthesis(remove_trailing_brackets(s))
}

remove_trailing_brackets(s) = clean if {
	no_space = trim_space(s)
	no_left = trim_left(no_space, "{")
	no_right = trim_right(no_left, "}")
	clean = trim_space(no_right)
}

clean_left(s) = clean if {
	no_space = trim_space(s)
	no_parenthesis = trim_left(no_space, "(")
	clean = trim_space(no_parenthesis)
}

clean_right(s) = clean if {
	no_space = trim_space(s)
	no_parenthesis = trim_right(no_space, ")")
	clean = trim_space(no_parenthesis)
}

remove_spaces_between_parenthesis(s) = new_string if {
	new_string = regex.replace(regex.replace(s, "\\)\\s+\\)", "))"), "\\(\\s+\\(", "((")
} else = s

complex_expression(op, expressions) = {
	"operator": op,
	"expressions": expressions,
}

simple_expression(left, op, right) = {
	"left": left,
	"operator": op,
	"right": right,
}

count_parenthesis(l) = val if {
	l == "("
	val = 1
}

count_parenthesis(l) = val if {
	l == ")"
	val = -1
}

count_parenthesis(l) = val if {
	l != "("
	l != ")"
	val = 0
}

is_parenthesis_balanced(s) if {
	symbols = split(s, "")

	# First convert everything to 1, 0 or -1
	values = [vals |
		symbol := symbols[_]
		vals := count_parenthesis(symbol)
	]

	# Then sum every sequential sub set of the array
	# so for an array of [1, 0, -1], it wil check the sums of:
	#   [1]
	#   [1, 0]
	#   [1, 0, -1]
	every index, _ in values {
		subset = [sub |
			values[i]
			i <= index # Filter by only indexes lesser than i
			sub = values[i]
		]

		sum(subset) >= 0
	}

	sum(values) == 0 # the total sum must be 0
} else = false
