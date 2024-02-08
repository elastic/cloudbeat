package compliance.policy.aws_cloudtrail.pattern

import future.keywords.if

filter_1 = {"FilterPattern": "filter_1", "FilterName": "filter_1"}

filter_2 = {"FilterPattern": "filter_2", "FilterName": "filter_2"}

pattern_1 = "filter_1"

pattern_2 = "filter_2"

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

test_parse if {
	is_equal(
		"simple expression",
		parse_expression("{$.eventName=DeleteGroupPolicy}"),
		simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"simple expression with spaces",
		parse_expression("{   $.eventName = DeleteGroupPolicy   }"),
		simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"simple expression with spaces in the middle",
		parse_expression("{   $. eventName = DeleteGroupPolicy   }"),
		simple_expression("$. eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"simple expression with string",
		parse_expression("{   $. eventName = \" String string string  \" }"),
		simple_expression("$. eventName", "=", "\" String string string  \""),
	)

	is_equal(
		"simple expression 'different' comparator",
		parse_expression("{   $. eventName != \" String string string  \" }"),
		simple_expression("$. eventName", "!=", "\" String string string  \""),
	)

	is_equal(
		"simple expression 'notExists' comparator",
		parse_expression("{   $.eventName NOT EXISTS }"),
		simple_expression("$.eventName", "NOT EXISTS", ""),
	)

	is_equal(
		"simple expression with parenthesis",
		parse_expression("{($.eventName=DeleteGroupPolicy)}"),
		simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"simple expression with multiple parenthesis",
		parse_expression("{(($.eventName=DeleteGroupPolicy))}"),
		simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"simple expression with parenthesis and spaces",
		parse_expression("{   (   $.eventName  =   DeleteGroupPolicy )   }"),
		simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	)

	is_equal(
		"complex expression 2 expressions AND operator",
		parse_expression("{$.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS}"),
		complex_expression("&&", [
			simple_expression("$.userIdentity.type", "=", "\"Root\""),
			simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
		]),
	)

	is_equal(
		"complex expression 2 expressions OR operator",
		parse_expression("{$.userIdentity.type = \"Root\" || $.userIdentity.invokedBy NOT EXISTS}"),
		complex_expression("||", [
			simple_expression("$.userIdentity.type", "=", "\"Root\""),
			simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
		]),
	)

	is_equal(
		"complex expression 2 expressions with outer parenthesis",
		parse_expression("{($.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS)}"),
		complex_expression("&&", [
			simple_expression("$.userIdentity.type", "=", "\"Root\""),
			simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
		]),
	)

	is_equal(
		"complex expression with parenthesis per simple expression",
		parse_expression("{($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"),
		complex_expression("||", [
			simple_expression("$.errorCode", "=", "\"*UnauthorizedOperation\""),
			simple_expression("$.errorCode", "=", "\"AccessDenied*\""),
			simple_expression("$.sourceIPAddress", "!=", "\"delivery.logs.amazonaws.com\""),
			simple_expression("$.eventName", "!=", "\"HeadBucket\""),
		]),
	)

	is_equal(
		"complex expression 3 expressions",
		parse_expression("{$.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }"),
		complex_expression("&&", [
			simple_expression("$.userIdentity.type", "=", "\"Root\""),
			simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
			simple_expression("$.eventType", "!=", "\"AwsServiceEvent\""),
		]),
	)

	is_equal(
		"complex expression 2 logical operators AND main OR sub",
		parse_expression("{($.eventSource = kms.amazonaws.com) && (($.eventName=DisableKey)||($.eventName=ScheduleKeyDeletion)) }"),
		complex_expression("&&", [
			simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
			complex_expression("||", [
				simple_expression("$.eventName", "=", "DisableKey"),
				simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
			]),
		]),
	)

	is_equal(
		"complex expression 2 logical operators || main OR sub",
		parse_expression("{($.eventSource = kms.amazonaws.com) || (($.eventName=DisableKey)&&($.eventName=ScheduleKeyDeletion)) }"),
		complex_expression("||", [
			simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
			complex_expression("&&", [
				simple_expression("$.eventName", "=", "DisableKey"),
				simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
			]),
		]),
	)

	is_equal(
		"complex expression 2 logical operators AND main AND sub",
		parse_expression("{($.eventSource = kms.amazonaws.com) && (($.eventName=DisableKey)&&($.eventName=ScheduleKeyDeletion)) }"),
		complex_expression("&&", [
			simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
			simple_expression("$.eventName", "=", "DisableKey"),
			simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
		]),
	)

	is_equal(
		"sub expression first",
		parse_expression("{ (($.eventName=DisableKey)||($.eventName=ScheduleKeyDeletion)) && ($.eventSource = kms.amazonaws.com) }"),
		complex_expression("&&", [
			complex_expression("||", [
				simple_expression("$.eventName", "=", "DisableKey"),
				simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
			]),
			simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
		]),
	)

	is_equal(
		"error on expression alternating logical operators",
		parse_expression("{($.eventSource = kms.amazonaws.com) && ($.eventName=DisableKey) || ($.eventName=ScheduleKeyDeletion)}"),
		{"error": "Could not parse expression"},
	)

	is_equal(
		"error on 4 layers deep expression",
		parse_expression("{((a=b) && ((c=d) || ((e=f) && (g!=h || (i=j)))))}"),
		{"error": "Could not parse expression"},
	)

	is_equal(
		"error on broken parenthesis and spaces",
		parse_expression("{   (   $.eventName  =   DeleteGroupPolicy ))   }"),
		{"error": "Could not parse expression"},
	)

	is_equal(
		"error on double operators (double equals)",
		parse_expression("{   $.eventName == a }"),
		{"error": "Could not parse expression"},
	)

	is_equal(
		"expect weird behaviour on double operators (different and equals)",
		parse_expression("{   $.eventName !== a }"),
		simple_expression("$.eventName", "!=", "= a"),
	)

	is_equal(
		"error on double operators (after expression)",
		parse_expression("{   $.eventName != a !=}"),
		{"error": "Could not parse expression"},
	)
}

test_is_parenthesis_balanced if {
	is_equal("simple parenthesis", is_parenthesis_balanced("()"), true)
	is_equal("broken parenthesis", is_parenthesis_balanced(")("), false)
	is_equal("broken parenthesis 2", is_parenthesis_balanced("())"), false)
	is_equal("broken parenthesis 3", is_parenthesis_balanced(")())"), false)
	is_equal("simple parenthesis with content", is_parenthesis_balanced("(a=b)"), true)
	is_equal("sub parenthesis", is_parenthesis_balanced("(a=b (sadasdasd))"), true)
	is_equal("broken sub parenthesis", is_parenthesis_balanced("(a=b (sadasdasd) ))"), false)
	is_equal("multiple expressions", is_parenthesis_balanced("()()()"), true)
	is_equal("multiple sub expressions", is_parenthesis_balanced("(()()(()()(())))"), true)
	is_equal("only opening", is_parenthesis_balanced("("), false)
	is_equal("only closing", is_parenthesis_balanced(")"), false)
}

test_is_main_operator_and if {
	is_equal("TRUE no parenthesis", is_main_operator_and("a=b && (c=d || e=f)"), true)
	is_equal("TRUE with parenthesis", is_main_operator_and("(a=b && (c=d || e=f))"), true)
	is_equal("TRUE with double parenthesis", is_main_operator_and("(a=b && ((c=d) || (e=f)))"), true)
	is_equal("TRUE with double parenthesis twice", is_main_operator_and("((a=b) && ((c=d) || (e=f)))"), true)
	is_equal("TRUE with expressions on both sides and double parenthesis", is_main_operator_and("((a=b) && ((c=d) || (e=f)) && (g=h))"), true)
	is_equal("TRUE main operator on the other side", is_main_operator_and("((c=d || e=f) && (a=b))"), true)
	is_equal("TRUE non wrapped expression", is_main_operator_and("(a = b) && ((c=d)||(e=f))"), true)
	is_equal("FALSE no parenthesis", is_main_operator_and("a=b || (c=d && e=f)"), false)
	is_equal("FALSE with parenthesis", is_main_operator_and("(a=b || (c=d && e=f))"), false)
	is_equal("FALSE with double parenthesis", is_main_operator_and("(a=b || ((c=d) && (e=f)))"), false)
	is_equal("FALSE with double parenthesis twice", is_main_operator_and("((a=b) || ((c=d) && (e=f)))"), false)
	is_equal("FALSE with expressions on both sides", is_main_operator_and("((a=b) || ((c=d) && (e=f)) && g=h)"), false)
	is_equal("FALSE with expressions on both sides and double parenthesis", is_main_operator_and("((a=b) || ((c=d) && (e=f)) && (g=h))"), false)
	is_equal("FALSE main operator on the other side", is_main_operator_and("((c=d && e=f) || (a=b))"), false)
	is_equal("FALSE non wrapped expression", is_main_operator_and("(a = b) || ((c=d)&&(e=f))"), false)
}

test_is_main_operator_or if {
	is_equal("TRUE no parenthesis", is_main_operator_or("a=b || (c=d && e=f)"), true)
	is_equal("TRUE with parenthesis", is_main_operator_or("(a=b || (c=d && e=f))"), true)
	is_equal("TRUE with double parenthesis", is_main_operator_or("(a=b || ((c=d) && (e=f)))"), true)
	is_equal("TRUE with double parenthesis twice", is_main_operator_or("((a=b) || ((c=d) && (e=f)))"), true)
	is_equal("TRUE with expressions on both sides and double parenthesis", is_main_operator_or("((a=b) || ((c=d) && (e=f)) || (g=h))"), true)
	is_equal("TRUE non wrapped expression", is_main_operator_or("(a = b) || ((c=d)&&(e=f))"), true)
	is_equal("TRUE main operator on the other side", is_main_operator_or("((c=d && e=f) || (a=b))"), true)
	is_equal("FALSE no parenthesis", is_main_operator_or("a=b && (c=d || e=f)"), false)
	is_equal("FALSE with parenthesis", is_main_operator_or("(a=b && (c=d || e=f))"), false)
	is_equal("FALSE with double parenthesis", is_main_operator_or("(a=b && ((c=d) || (e=f)))"), false)
	is_equal("FALSE with double parenthesis twice", is_main_operator_or("((a=b) && ((c=d) || (e=f)))"), false)
	is_equal("FALSE with expressions on both sides", is_main_operator_or("((a=b) && ((c=d) || (e=f)) && g=h)"), false)
	is_equal("FALSE with expressions on both sides and double parenthesis", is_main_operator_or("((a=b) && ((c=d) || (e=f)) && (g=h))"), false)
	is_equal("FALSE main operator on the other side", is_main_operator_or("((c=d || e=f) && (a=b))"), false)
	is_equal("FALSE non wrapped expression", is_main_operator_or("(a = b) && ((c=d)||(e=f))"), false)
}

test_find_complexity if {
	is_equal("Test simple expression", find_complexity("a = b"), "SIMPLE")
	is_equal("Test simple expression parenthesis", find_complexity("(a = b)"), "SIMPLE")
	is_equal("Test complex expression 1", find_complexity("a = b || b = c"), "COMPLEX")
	is_equal("Test complex expression 2", find_complexity("a = b || (b = c)"), "COMPLEX")
	is_equal("Test complex expression 3", find_complexity("(a = b || (b = c))"), "COMPLEX")
	is_equal("Test complex expression 4", find_complexity("(a = b || (b = c) || (d = e))"), "COMPLEX")
	is_equal("Test complex expression 5", find_complexity("(a = b || (b = c && e = f))"), "COMPLEX")
	is_equal("Invalid no parenthesis", find_complexity("a = b || b = c && e = f"), "INVALID")
	is_equal("Invalid with parenthesis", find_complexity("(a = b || b = c && e = f)"), "INVALID")
	is_equal("Invalid with double parenthesis", find_complexity("(a = b || (b = c) && e = f)"), "INVALID")
	is_equal("Invalid with valid sub expression", find_complexity("(a = b || (b = c && a = c) && e = f)"), "INVALID")
	is_equal("Invalid with valid sub expression more parenthesis", find_complexity("((a=b) || ((c=d) && (e=f)) && g=h)"), "INVALID")
}

test_count_parenthesis_levels if {
	is_equal("0 level simple", count_parenthesis_levels("a=b"), 0)
	is_equal("0 level double", count_parenthesis_levels("a=b || c=d"), 0)
	is_equal("1 level simple", count_parenthesis_levels("(a=b)"), 1)
	is_equal("1 level double", count_parenthesis_levels("(a=b) || (c=d)"), 1)
	is_equal("2 level simple", count_parenthesis_levels("((a=b))"), 2)
	is_equal("2 level double", count_parenthesis_levels("((a=b) || (c=d))"), 2)
	is_equal("3 level tripple", count_parenthesis_levels("((a=b) || ((c=d) && (e=d)))"), 3)
}

test_count_complexity_levels if {
	is_equal("0 level simple", count_complexity_levels("a=b"), 0)
	is_equal("0 level simple 1 parentehesis", count_complexity_levels("(a=b)"), 0)
	is_equal("0 level simple 2 parentehesis", count_complexity_levels("((a=b))"), 0)
	is_equal("0 level simple a lot of parentehesis", count_complexity_levels("(((((((((a=b)))))))))"), 0)

	is_equal("1 level simple", count_complexity_levels("a=b && c=d"), 1)
	is_equal("1 level simple 1 parentehesis", count_complexity_levels("(a=b && c=d)"), 1)
	is_equal("1 level simple 1 parentehesis", count_complexity_levels("(a=b) && (c=d)"), 1)
	is_equal("1 level simple 2 parentehesis", count_complexity_levels("((a=b) && (c=d))"), 1)
	is_equal("1 level simple a lot of parentehesis", count_complexity_levels("(((((((((a=b) && (c=d) && (g=h)))))))))"), 1)

	is_equal("2 level simple", count_complexity_levels("a=b && (c=d || e=f)"), 2)
	is_equal("2 level simple 1 parentehesis", count_complexity_levels("(a=b && (c=d || e=f))"), 2)
	is_equal("2 level simple 1 parentehesis", count_complexity_levels("(a=b) && ((c=d || e=f))"), 2)
	is_equal("2 level simple 2 parentehesis", count_complexity_levels("((a=b) && ((c=d || e=f || (g=h))))"), 2)
	is_equal("2 level simple a lot of parentehesis", count_complexity_levels("(((((((((a=b) && ((c=d || ((e=f))))))))))))"), 2)

	is_equal("3 level simple", count_complexity_levels("a=b && (c=d || (e=f && (g=h)))"), 3)

	is_equal("4 level", count_complexity_levels("((a=b) && ((c=d) || ((e=f) && (g!=h || (i=j)))))"), 4)
}

is_equal(_, actual, want) if {
	actual == want
}

is_equal(desc, actual, want) if {
	actual != want
	print("--- Test [", desc, "] failed because:")
	print("WANT:    ", want)
	print("ACTUAL:  ", actual)
	false
}
