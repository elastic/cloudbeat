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

    # TODO BROKEN TEST NEED TO FIND A WAY TO INDENTIFY ALTERNATING LOGICAL OPERATORS
	is_equal(
		"error on expression alternating logical operators",
		parse_expression("{($.eventSource = kms.amazonaws.com) && ($.eventName=DisableKey) || ($.eventName=ScheduleKeyDeletion)}"),
		complex_expression("&&", [
			simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
			complex_expression("||", [
				simple_expression("$.eventName", "=", "DisableKey"),
				simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
			]),
		]),
	)
}

test_remove_spaces_between_parenthesis if {
	is_equal("no parenthesis", remove_spaces_between_parenthesis("a=b"), "a=b")
	is_equal("simple parenthesis", remove_spaces_between_parenthesis("(a=b)"), "(a=b)")
	is_equal("double parenthesis", remove_spaces_between_parenthesis("((a=b))"), "((a=b))")
	is_equal("double parenthesis with space", remove_spaces_between_parenthesis("( (a=b) )"), "((a=b))")
	is_equal("double parenthesis with multiple spaces", remove_spaces_between_parenthesis("(\t    (a=b) )"), "((a=b))")
	is_equal("double parenthesis with trailing spaces", remove_spaces_between_parenthesis("  (\t    (a=b) )  "), "  ((a=b))  ")
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
	is_equal("TRUE with expressions on both sides", is_main_operator_and("((a=b) && ((c=d) || (e=f)) && g=h)"), true)
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
	is_equal("TRUE with expressions on both sides", is_main_operator_or("((a=b) || ((c=d) && (e=f)) && g=h)"), true)
	is_equal("TRUE with expressions on both sides and double parenthesis", is_main_operator_or("((a=b) || ((c=d) && (e=f)) && (g=h))"), true)
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
