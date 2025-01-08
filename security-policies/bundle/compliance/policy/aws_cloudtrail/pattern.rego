package compliance.policy.aws_cloudtrail.pattern

import future.keywords.every
import future.keywords.if
import future.keywords.in

# get a filter from a trail has at least one metric filter pattern that matches at least one pattern
get_filter_matched_to_pattern(trail, patterns) := name if {
	some i, j
	filter := trail.MetricFilters[i]
	pattern := patterns[j]
	expressions_equivalent(filter.ParsedFilterPattern, pattern)
	name := filter.FilterName
} else := ""

complex_expression(op, expressions) := {
	"ComparisonOperator": "",
	"Complex": true,
	"Expressions": expressions,
	"Left": "",
	"LogicalOperator": op,
	"Right": "",
	"Simple": false,
}

simple_expression(left, op, right) := {
	"ComparisonOperator": op,
	"Complex": false,
	"Expressions": null,
	"Left": left,
	"LogicalOperator": "",
	"Right": right,
	"Simple": true,
}

# Known limitations on checking expressions equivalence:
#   - It supports only two levels deep expressions (2 levels are as deep as our uses cases go)
default expressions_equivalent(_, _) := false

expressions_equivalent(exp1, exp2) if {
	compare_simple_expressions(exp1, exp2)
}

expressions_equivalent(exp1, exp2) if {
	compare_complex_expressions(exp1, exp2)
}

compare_simple_expressions(exp1, exp2) if {
	exp1.Simple
	exp2.Simple
	exp1.Left == exp2.Left
	exp1.ComparisonOperator == exp2.ComparisonOperator
	exp1.Right == exp2.Right
}

compare_simple_expressions(exp1, exp2) if {
	exp1.Simple
	exp2.Simple
	exp1.Left == exp2.Right
	exp1.ComparisonOperator == exp2.ComparisonOperator
	exp1.Right == exp2.Left
}

compare_complex_expressions(exp1, exp2) if {
	exp1.Complex
	exp2.Complex
	exp1.LogicalOperator == exp2.LogicalOperator
	count(exp1.Expressions) == count(exp2.Expressions)

	every subExp1 in exp1.Expressions {
		some subExp2 in exp2.Expressions
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
	exp1.Complex
	exp2.Complex
	exp1.LogicalOperator == exp2.LogicalOperator
	count(exp1.Expressions) == count(exp2.Expressions)

	every subExp1 in exp1.Expressions {
		some subExp2 in exp2.Expressions
		compare_simple_expressions(subExp1, subExp2)
	}
}
