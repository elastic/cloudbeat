package compliance.policy.aws_cloudtrail.pattern

import future.keywords.every
import future.keywords.if
import future.keywords.in

# get a filter from a trail has at least one metric filter pattern that matches at least one pattern
get_filter_matched_to_pattern(trail, patterns) = name if {
	some i, j
	filter := trail.MetricFilters[i]
	pattern := patterns[j]
	expressions_equivalent(filter.FilterPattern, pattern)
	name := filter.FilterName
} else = ""

complex_expression(op, expressions) = {
	"complex": true,
	"operator": op,
	"expressions": expressions,
}

simple_expression(left, op, right) = {
	"simple": true,
	"left": left,
	"operator": op,
	"right": right,
}

default expressions_equivalent(expression1, expression2) = false

expressions_equivalent(str1, str2) if {
	exp1 = parse_expression(str1)
	not exp1.error
	exp2 = parse_expression(str2)
	not exp2.error
	compare_simple_expressions(exp1, exp2)
}

expressions_equivalent(str1, str2) if {
	exp1 = parse_expression(str1)
	not exp1.error
	exp2 = parse_expression(str2)
	not exp2.error
	compare_complex_expressions(exp1, exp2)
}

compare_simple_expressions(exp1, exp2) if {
	exp1.simple
	exp2.simple
	exp1.left == exp2.left
	exp1.operator == exp2.operator
	exp1.right == exp2.right
}

compare_simple_expressions(exp1, exp2) if {
	exp1.simple
	exp2.simple
	exp1.left == exp2.right
	exp1.operator == exp2.operator
	exp1.right == exp2.left
}

compare_complex_expressions(exp1, exp2) if {
	exp1.complex
	exp2.complex
	exp1.operator == exp2.operator
	count(exp1.expressions) == count(exp2.expressions)

	every subExp1 in exp1.expressions {
		some subExp2 in exp2.expressions
		compare_expressions_second_level(subExp1, subExp2)
	}
}

compare_expressions_second_level(exp1, exp2) if {
	compare_simple_expressions(exp1, exp2)
}

compare_expressions_second_level(exp1, exp2) if {
	compare_complex_expressions_second_level(exp1, exp2)
}

compare_complex_expressions_second_level(exp1, exp2) if {
	exp1.complex
	exp2.complex
	exp1.operator == exp2.operator
	count(exp1.expressions) == count(exp2.expressions)

	every subExp1 in exp1.expressions {
		some subExp2 in exp2.expressions
		compare_simple_expressions(subExp1, subExp2)
	}
}

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
	is_main_operator_and(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Process
	expression = parse_two_operators_expression(clean_s, "&&")
}

parse_expression(s) = expression if { # handle complex 2 operator (|| as main operator)
	# Filter
	is_complex_expression(s)
	has_and_operator(s)
	has_or_operator(s)
	is_main_operator_or(s)

	# Clean
	clean_s = clean_full_expression(s)

	# Process
	expression = parse_two_operators_expression(clean_s, "||")
}

is_simple_expression(s) if {
	find_complexity(s) == "SIMPLE"
}

is_complex_expression(s) if {
	find_complexity(s) == "COMPLEX"
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

is_sub_expression_simple_or_has_or_operator(s) if {
	is_parenthesis_balanced(s)
	has_or_operator(s)
}

is_sub_expression_simple_or_has_or_operator(s) if {
	is_simple_expression(s)
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

	# Parse
	expression = parse_simple_expression(s)
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

	# Parse
	expression = parse_simple_expression(s)
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
	clean = remove_space_between_parenthesis(remove_trailing_brackets(s))
}

remove_space_between_parenthesis(s) = clean if {
	clean_l = regex.replace(s, "(?:\\()\\s+\\(", "((")
	clean = regex.replace(clean_l, "\\)\\s+\\)", "))")
} else = s

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

default count_parenthesis(l) = 0

count_parenthesis(l) = val if {
	l == "("
	val = 1
}

count_parenthesis(l) = val if {
	l == ")"
	val = -1
}

default is_parenthesis_balanced(s) = false

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
}

count_parenthesis_levels(s) = levels if {
	symbols = split(s, "")

	# First convert everything to 1, 0 or -1
	values = [vals |
		symbol := symbols[_]
		vals := count_parenthesis(symbol)
	]

	# Sum subsets of values to find the levels. e.g
	# Given the symbols
	#   [ (, a, =, b, ) ]
	# This is converted to values
	#   [ 1, 0, 0, 0, -1 ]
	# When performing the sum for each index of values from 0 until the index, we get the array
	#   [ 1, 1, 1, 1, 0 ]
	# That because for each index the following sum is performed:
	#   idx 0 -> sum([ 1 ]) = 1
	#   idx 1 -> sum([ 1, 0 ]) = 1
	#   idx 2 -> sum([ 1, 0, 0 ]) = 1
	#   idx 3 -> sum([ 1, 0, 0, 0 ]) = 1
	#   idx 4 -> sum([ 1, 0, 0, 0, -1 ]) = 0

	subset = [sub |
		values[i]
		sub = sum(array.slice(values, 0, i + 1))
	]

	levels = max(subset)
}

default count_complexity_symbols(l) = 0

# We use prime numbers to make sure that the complexity division won't have colisions
# We won't ever get there, but good to know, if we ever get to 23 levels, colisions will exist

and_complexity_score := 29 # prime number

count_complexity_symbols(l) = val if {
	l == "&"
	val = and_complexity_score
}

or_complexity_score := 23 # prime number

count_complexity_symbols(l) = val if {
	l == "|"
	val = or_complexity_score
}

get_leveled_operators(s) = leveled_operators if {
	symbols = split(s, "")

	# First grab parentehsis
	parenthesis = [vals |
		symbol := symbols[_]
		vals := count_parenthesis(symbol)
	]

	# Reduce parenthesis to levels
	levels = [sub |
		parenthesis[i]
		sub = sum(array.slice(parenthesis, 0, i + 1))
	]

	# Then find symbols times their level (in a set for efficiency since order don't matter here)
	leveled_operators = {vals |
		symbol := symbols[i]
		vals := count_complexity_symbols(symbol) * (levels[i] + 1)
	}
}

default find_complexity(s) = "INVALID"

max_allowed_complexity := 2

find_complexity(s) = complexity if {
	# Filter
	count_complexity_levels(s) <= max_allowed_complexity
	is_parenthesis_balanced(s)

	leveled_operators = get_leveled_operators(s)

	# There are no symbol
	leveled_operators = {0}
	complexity = "SIMPLE"
}

find_complexity(s) = complexity if {
	# Filter
	count_complexity_levels(s) <= max_allowed_complexity
	is_parenthesis_balanced(s)

	leveled_operators = get_leveled_operators(s)

	# There are symbols
	leveled_operators != {0}

	# If in any level we have symbols twice it's invalid
	every level in numbers.range(1, count_parenthesis_levels(s) + 1) {
		level_not_contains_double_operators(level, leveled_operators)
	}

	complexity = "COMPLEX"
}

default level_not_contains_double_operators(level, leveled_operators) = false

# doesn't contain at all
level_not_contains_double_operators(level, leveled_operators) if {
	not leveled_operators[level * and_complexity_score]
	not leveled_operators[level * or_complexity_score]
}

# doesn't contain only AND
level_not_contains_double_operators(level, leveled_operators) if {
	leveled_operators[level * and_complexity_score]
	not leveled_operators[level * or_complexity_score]
}

# doesn't contain only OR
level_not_contains_double_operators(level, leveled_operators) if {
	leveled_operators[level * or_complexity_score]
	not leveled_operators[level * and_complexity_score]
}

default is_main_operator_and(s) = false

is_main_operator_and(s) if {
	leveled_operators = get_leveled_operators(s)
	levels = numbers.range(1, count_parenthesis_levels(s) + 1)

	# Check what is the operator of each level
	main_operator_per_level = [main |
		level = levels[_]
		main = main_operator_in_level(level, leveled_operators)
	]

	levels_with_and_as_main = [level |
		op = main_operator_per_level[level]
		op == "AND"
	]

	levels_with_or_as_main = [level |
		op = main_operator_per_level[level]
		op == "OR"
	]

	min(levels_with_and_as_main) < min(levels_with_or_as_main)
}

default is_main_operator_or(s) = false

is_main_operator_or(s) if {
	leveled_operators = get_leveled_operators(s)
	levels = numbers.range(1, count_parenthesis_levels(s) + 1)

	# Check what is the operator of each level
	main_operator_per_level = [main |
		level = levels[_]
		main = main_operator_in_level(level, leveled_operators)
	]

	levels_with_and_as_main = [level |
		op = main_operator_per_level[level]
		op == "AND"
	]

	levels_with_or_as_main = [level |
		op = main_operator_per_level[level]
		op == "OR"
	]

	min(levels_with_and_as_main) > min(levels_with_or_as_main)
}

default main_operator_in_level(level, leveled_operators) = ""

main_operator_in_level(level, leveled_operators) = main if {
	leveled_operators[level * and_complexity_score]
	not leveled_operators[level * or_complexity_score]
	main = "AND"
}

main_operator_in_level(level, leveled_operators) = main if {
	leveled_operators[level * or_complexity_score]
	not leveled_operators[level * and_complexity_score]
	main = "OR"
}

count_complexity_levels(s) = valid_levels if {
	leveled_operators = get_leveled_operators(s)
	levels = numbers.range(1, count_parenthesis_levels(s) + 1)

	# Check what is the operator of each level
	main_operator_per_level = [main |
		level = levels[_]
		main = main_operator_in_level(level, leveled_operators)
		main != ""
	]

	valid_levels = count(main_operator_per_level)
}
