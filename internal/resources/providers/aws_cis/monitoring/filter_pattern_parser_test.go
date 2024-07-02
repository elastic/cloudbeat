// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package monitoring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFilterPattern(t *testing.T) {
	cases := map[string]struct {
		in  string
		out MetricFilterPattern
		err error
	}{
		"simple expression": {
			in:  "{$.eventName=DeleteGroupPolicy}",
			out: newSimpleExpression("$.eventName", coEqual, "DeleteGroupPolicy"),
		},
		"simple expression with spaces": {
			in:  "{   $.eventName = DeleteGroupPolicy   }",
			out: newSimpleExpression("$.eventName", coEqual, "DeleteGroupPolicy"),
		},
		"simple expression with spaces in the middle": {
			in:  "{   $. eventName = DeleteGroupPolicy   }",
			out: newSimpleExpression("$. eventName", coEqual, "DeleteGroupPolicy"),
		},
		"simple expression with string": {
			in:  "{   $. eventName = \" String string string  \" }",
			out: newSimpleExpression("$. eventName", coEqual, "\" String string string  \""),
		},
		"simple expression 'different' comparator": {
			in:  "{   $. eventName != \" String string string  \" }",
			out: newSimpleExpression("$. eventName", coNotEqual, "\" String string string  \""),
		},
		"simple expression 'notExists' comparator": {
			in:  "{   $.eventName NOT EXISTS }",
			out: newSimpleExpression("$.eventName", coNotExists, ""),
		},
		"simple expression with parenthesis": {
			in:  "{($.eventName=DeleteGroupPolicy)}",
			out: newSimpleExpression("$.eventName", coEqual, "DeleteGroupPolicy"),
		},
		"simple expression with multiple parenthesis": {
			in:  "{(((($.eventName=DeleteGroupPolicy))))}",
			out: newSimpleExpression("$.eventName", coEqual, "DeleteGroupPolicy"),
		},
		"simple expression with parenthesis and spaces": {
			in:  "{   (   $.eventName  =   DeleteGroupPolicy )   }",
			out: newSimpleExpression("$.eventName", coEqual, "DeleteGroupPolicy"),
		},
		"error on broken parenthesis and spaces": {
			in:  "{   (   $.eventName  =   DeleteGroupPolicy ))   }",
			err: errors.New("broken parenthesis"),
		},
		"error on double operators (double equals)": {
			in:  "{   $.eventName == a }",
			err: errors.New("got multiple comparison operators"),
		},
		"error on double operators (different and equals)": {
			in:  "{   $.eventName !== a }",
			err: errors.New("got multiple comparison operators"),
		},
		"error on double operators (after expression)": {
			in:  "{   $.eventName != a !=}",
			err: errors.New("got multiple comparison operators"),
		},
		"complex expression 2 expressions": {
			in: "{$.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS}",
			out: newComplexExpression(loAnd,
				newSimpleExpression("$.userIdentity.type", coEqual, "\"Root\""),
				newSimpleExpression("$.userIdentity.invokedBy", coNotExists, "")),
		},
		"complex expression 2 expressions with outer parenthesis": {
			in: "{($.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS)}",
			out: newComplexExpression(loAnd,
				newSimpleExpression("$.userIdentity.type", coEqual, "\"Root\""),
				newSimpleExpression("$.userIdentity.invokedBy", coNotExists, "")),
		},
		"complex expression with parenthesis per simple expression": {
			in: "{($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			out: newComplexExpression(loOr,
				newSimpleExpression("$.errorCode", coEqual, "\"*UnauthorizedOperation\""),
				newSimpleExpression("$.errorCode", coEqual, "\"AccessDenied*\""),
				newSimpleExpression("$.sourceIPAddress", coNotEqual, "\"delivery.logs.amazonaws.com\""),
				newSimpleExpression("$.eventName", coNotEqual, "\"HeadBucket\""),
			),
		},
		"complex expression 3 expressions": {
			in: "{$.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }",
			out: newComplexExpression(loAnd,
				newSimpleExpression("$.userIdentity.type", coEqual, "\"Root\""),
				newSimpleExpression("$.userIdentity.invokedBy", coNotExists, ""),
				newSimpleExpression("$.eventType", coNotEqual, "\"AwsServiceEvent\""),
			),
		},
		"complex expression 2 logical operators": {
			in: "{($.eventSource = kms.amazonaws.com) && (($.eventName=DisableKey)||($.eventName=ScheduleKeyDeletion)) }",
			out: newComplexExpression(loAnd,
				newSimpleExpression("$.eventSource", coEqual, "kms.amazonaws.com"),
				newComplexExpression(loOr,
					newSimpleExpression("$.eventName", coEqual, "DisableKey"),
					newSimpleExpression("$.eventName", coEqual, "ScheduleKeyDeletion"),
				),
			),
		},
		"sub expression first": {
			in: "{ (($.eventName=DisableKey)||($.eventName=ScheduleKeyDeletion)) && ($.eventSource = kms.amazonaws.com) }",
			out: newComplexExpression(loAnd,
				newComplexExpression(loOr,
					newSimpleExpression("$.eventName", coEqual, "DisableKey"),
					newSimpleExpression("$.eventName", coEqual, "ScheduleKeyDeletion"),
				),
				newSimpleExpression("$.eventSource", coEqual, "kms.amazonaws.com"),
			),
		},
		"error on complex expression alternating logical operators": {
			in:  "{($.eventSource = kms.amazonaws.com) && ($.eventName=DisableKey) || ($.eventName=ScheduleKeyDeletion)}",
			err: errors.New("not supported comparison with alternating logical operators"),
		},
		"4 layers deep expression": {
			in: "{((a=b) && ((c=d) || ((e=f) && (g!=h || (i=j)))))}",
			out: newComplexExpression(loAnd,
				newSimpleExpression("a", coEqual, "b"),
				newComplexExpression(loOr,
					newSimpleExpression("c", coEqual, "d"),
					newComplexExpression(loAnd,
						newSimpleExpression("e", coEqual, "f"),
						newComplexExpression(loOr,
							newSimpleExpression("g", coNotEqual, "h"),
							newSimpleExpression("i", coEqual, "j"),
						),
					),
				),
			),
		},
		"error on too deep expression": {
			in:  "{((((((((((((a=b)&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))&&(a=b))}",
			err: errors.New("max depth reached, can't parse this expression"),
		},

		"Long expression": {
			in: "{ ($.eventSource = organizations.amazonaws.com) && (($.eventName = \"AttachPolicy\") || ($.eventName = \"CreateAccount\") || ($.eventName = \"CreateOrganizationalUnit\") || ($.eventName = \"CreatePolicy\") || ($.eventName = \"DeclineHandshake\") || ($.eventName = \"DeleteOrganization\") || ($.eventName = \"DeleteOrganizationalUnit\") || ($.eventName = \"DeletePolicy\") || ($.eventName = \"DetachPolicy\") || ($.eventName = \"DisablePolicyType\") || ($.eventName = \"EnablePolicyType\") || ($.eventName = \"InviteAccountToOrganization\") || ($.eventName = \"LeaveOrganization\") || ($.eventName = \"MoveAccount\") || ($.eventName = \"RemoveAccountFromOrganization\") || ($.eventName = \"AcceptHandshake\") ||  ($.eventName = \"UpdatePolicy\") || ($.eventName = \"UpdateOrganizationalUnit\")) }",
			out: newComplexExpression(loAnd,
				newSimpleExpression("$.eventSource", coEqual, "organizations.amazonaws.com"),
				newComplexExpression(loOr,
					newSimpleExpression("$.eventName", coEqual, "\"AttachPolicy\""),
					newSimpleExpression("$.eventName", coEqual, "\"CreateAccount\""),
					newSimpleExpression("$.eventName", coEqual, "\"CreateOrganizationalUnit\""),
					newSimpleExpression("$.eventName", coEqual, "\"CreatePolicy\""),
					newSimpleExpression("$.eventName", coEqual, "\"DeclineHandshake\""),
					newSimpleExpression("$.eventName", coEqual, "\"DeleteOrganization\""),
					newSimpleExpression("$.eventName", coEqual, "\"DeleteOrganizationalUnit\""),
					newSimpleExpression("$.eventName", coEqual, "\"DeletePolicy\""),
					newSimpleExpression("$.eventName", coEqual, "\"DetachPolicy\""),
					newSimpleExpression("$.eventName", coEqual, "\"DisablePolicyType\""),
					newSimpleExpression("$.eventName", coEqual, "\"EnablePolicyType\""),
					newSimpleExpression("$.eventName", coEqual, "\"InviteAccountToOrganization\""),
					newSimpleExpression("$.eventName", coEqual, "\"LeaveOrganization\""),
					newSimpleExpression("$.eventName", coEqual, "\"MoveAccount\""),
					newSimpleExpression("$.eventName", coEqual, "\"RemoveAccountFromOrganization\""),
					newSimpleExpression("$.eventName", coEqual, "\"AcceptHandshake\""),
					newSimpleExpression("$.eventName", coEqual, "\"UpdatePolicy\""),
					newSimpleExpression("$.eventName", coEqual, "\"UpdateOrganizationalUnit\""),
				),
			),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := parseFilterPattern(tc.in)
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.out, s)
		})
	}
}

/*
Simple replaces    	   36765	     31604 ns/op	   37096 B/op	      99 allocs/op
Regex replaces    	   16062	     78596 ns/op	   45002 B/op	     147 allocs/op
*/
func BenchmarkParseFilterPattern(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parseFilterPattern("{ ($.eventSource = organizations.amazonaws.com) && (($.eventName = \"AttachPolicy\") || ($.eventName = \"CreateAccount\") || ($.eventName = \"CreateOrganizationalUnit\") || ($.eventName = \"CreatePolicy\") || ($.eventName = \"DeclineHandshake\") || ($.eventName = \"DeleteOrganization\") || ($.eventName = \"DeleteOrganizationalUnit\") || ($.eventName = \"DeletePolicy\") || ($.eventName = \"DetachPolicy\") || ($.eventName = \"DisablePolicyType\") || ($.eventName = \"EnablePolicyType\") || ($.eventName = \"InviteAccountToOrganization\") || ($.eventName = \"LeaveOrganization\") || ($.eventName = \"MoveAccount\") || ($.eventName = \"RemoveAccountFromOrganization\") || ($.eventName = \"AcceptHandshake\") ||  ($.eventName = \"UpdatePolicy\") || ($.eventName = \"UpdateOrganizationalUnit\")) }")
		if err != nil {
			b.Fail()
		}
	}
}
