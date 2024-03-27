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
	"fmt"
	"strings"
)

const maxDepth = 5

type logicalOperator string
type comparisonOperator string

const (
	loAnd logicalOperator = "&&"
	loOr  logicalOperator = "||"

	coEqual     comparisonOperator = "="
	coNotEqual  comparisonOperator = "!="
	coNotExists comparisonOperator = "NOT EXISTS"
)

func listLogicalOperators() []logicalOperator {
	return []logicalOperator{loAnd, loOr}
}

func listComparisonOperators() []comparisonOperator {
	// This order must be kept because we need to check first different and then equals
	return []comparisonOperator{coNotExists, coNotEqual, coEqual}
}

// MetricFilterPattern is a union between simpleExpression and complexExpression
// because a metric filter pattern it can be either one or the other.
// The fields are never used in go layer, only in OPA layer, therefore the nolint comment
type MetricFilterPattern struct {
	simpleExpression  //nolint:unused
	complexExpression //nolint:unused
}

type simpleExpression struct {
	Simple             bool
	Left               string
	ComparisonOperator comparisonOperator
	Right              string
}

func newSimpleExpression(left string, op comparisonOperator, right string) MetricFilterPattern {
	return MetricFilterPattern{
		simpleExpression: simpleExpression{
			Simple:             true,
			Left:               left,
			ComparisonOperator: op,
			Right:              right,
		},
	}
}

type complexExpression struct {
	Complex         bool
	LogicalOperator logicalOperator
	Expressions     []MetricFilterPattern
}

func newComplexExpression(op logicalOperator, exps ...MetricFilterPattern) MetricFilterPattern {
	return MetricFilterPattern{
		complexExpression: complexExpression{
			Complex:         true,
			LogicalOperator: op,
			Expressions:     exps,
		},
	}
}

// parseFilterPattern receives a string with CloudWatch FilterPattern syntax and parses to an expressions
// Known limitations:
//   - It parses only 5 levels deep expressions. You can increase with maxDepth const. But we don't have that use case
//   - It doesn't properly handle reserved characters (=, !=, NOT EXISTS, &&, ||) inside strings. We don't have that use case yet
//   - It doesn't parse subsequent alternating Logical Operators. For example the expression `{a=b && c=d || e=f}` will
//     return an error. The expected way to have this working is by wrapping one of the 'sub-expressions' in parentheses
//     e.g `{(a=b && c=d) || e=f}` or `{a=b && (c=d || e=f)}`
func parseFilterPattern(s string) (MetricFilterPattern, error) {
	cleanS := cleanExpression(s)

	if strings.Count(s, "(") != strings.Count(s, ")") {
		return MetricFilterPattern{}, errors.New("broken parenthesis")
	}

	return safeParse(cleanS, 0)
}

// cleanExpression removes trailing  spaces and brackets from the expression
// Receives full expression e.g ` { a = b } `
// Returns clean expression e.g `a = b`
func cleanExpression(s string) string {
	// If receive for example the expression ` { a = b } `
	firstSpace := strings.TrimSpace(s)              // Removes outer spaces `{ a = b }`
	cleanLeft := strings.TrimLeft(firstSpace, "{")  // Removes left bracket ` a = b }`
	cleanRight := strings.TrimRight(cleanLeft, "}") // Removes right bracket ` a = b `
	totallyClean := strings.TrimSpace(cleanRight)   // Removes inner spaces `a = b`
	return totallyClean
}

// safeParse adds a depth parameter to parse and creates a lock to not parse too long expressions
// and prevent infinite loops
func safeParse(s string, depth int) (MetricFilterPattern, error) {
	if depth > maxDepth {
		return MetricFilterPattern{}, errors.New("max depth reached, can't parse this expression")
	}

	var logicalOp logicalOperator
	// Capacity is an estimate of a maximum average of expressions. This avoids resizing slices every time we add a new item
	expressions := make([]MetricFilterPattern, 0, 20)

	buf := strings.Builder{}
	buf.Grow(len(s)) // grow buffer to max amount of runes it will have

	pointer := 0
	for len(s) > pointer {
		r := rune(s[pointer])
		i := pointer
		pointer++

		if r == '(' { // If it's a parenthesis opening, resolve the parenthesis
			exp, closingParenthesisPos, err := resolveParenthesis(s, depth, i)
			if err != nil {
				return MetricFilterPattern{}, err
			}
			expressions = append(expressions, exp)
			pointer = closingParenthesisPos + i + 1 // move pointer to the end of what has been already processed
			continue
		}

		if _, err := buf.WriteRune(r); err != nil {
			return MetricFilterPattern{}, fmt.Errorf("could not write rune %v", err)
		}

		tmpString := buf.String()
		if contains, op := hasSuffixLogicalOp(tmpString); contains { // && or || marks the end of a Simple MetricFilterPattern
			if logicalOp == "" { // set the LogicalOperator of the MetricFilterPattern without overriding it
				logicalOp = op
			}

			if logicalOp != op {
				return MetricFilterPattern{}, errors.New("not supported comparison with alternating logical operators")
			}

			cleanSimpleExpression := strings.TrimSuffix(tmpString, string(op))

			var err error
			expressions, err = appendSimpleExpression(cleanSimpleExpression, expressions)
			if err != nil {
				return MetricFilterPattern{}, err
			}

			buf.Reset()
			buf.Grow(len(s) - i)
		}
	}

	expressions, err := appendSimpleExpression(buf.String(), expressions)
	if err != nil {
		return MetricFilterPattern{}, err
	}

	if len(expressions) == 1 { // unwrap Simple Expressions
		return expressions[0], nil
	}

	return newComplexExpression(logicalOp, expressions...), nil
}

func appendSimpleExpression(s string, expressions []MetricFilterPattern) ([]MetricFilterPattern, error) {
	expStr := strings.TrimSpace(s)
	// if the length is zero it means we had an already processed Complex Expressions (between parenthesis)
	if len(expStr) > 0 {
		exp, err := parseSimpleExpression(expStr)
		if err != nil {
			return nil, err
		}

		expressions = append(expressions, exp)
	}
	return expressions, nil
}

// resolveParenthesis find the matching parenthesis and safeParse the expression inside the parenthesis
// Returns the found MetricFilterPattern and closingParenthesisPos so the main algorithm
// knows where to move the pointer too
func resolveParenthesis(s string, depth int, i int) (MetricFilterPattern, int, error) {
	closingParenthesisPos := matchingParenthesisPos(s[i:])
	if closingParenthesisPos < 0 {
		return MetricFilterPattern{}, -1, errors.New("broken parenthesis")
	}

	subS := s[i+1 : closingParenthesisPos+i] // take what is inside matching parenthesis
	exp, err := safeParse(subS, depth+1)     // safe sub MetricFilterPattern
	if err != nil {
		return MetricFilterPattern{}, -1, err
	}
	return exp, closingParenthesisPos, nil
}

// matchingParenthesisPos find the closing parenthesis of the first opening parenthesis found in string
func matchingParenthesisPos(s string) int {
	parenthesisCount := 0
	for i, r := range s {
		switch r {
		case '(': // If an opening parenthesis appear add 1 to the count
			parenthesisCount++
		case ')': // If a closing parenthesis appear remove 1 to the count
			parenthesisCount--
			if parenthesisCount == 0 { // if count is 0 it means that we  just found the closing parenthesis position
				return i
			}
		default:
			continue
		}
	}

	return -1
}

// parseSimpleExpression receives a simple expression string e.g `a=b` and returns a simpleExpression
func parseSimpleExpression(s string) (MetricFilterPattern, error) {
	buf := strings.Builder{}
	buf.Grow(len(s))

	var left string
	var operator comparisonOperator
	foundOp := false

	for i, r := range s { // for each rune in string
		if buf.Len() == 0 && (r == ' ') { // ignore trailing spaces
			continue
		}

		// append the rune to the buffer
		if _, err := buf.WriteRune(r); err != nil {
			return MetricFilterPattern{}, fmt.Errorf("could not write rune %v", err)
		}

		tmpString := buf.String()
		// If the current buffer value has a Comparison ComparisonOperator as suffix, it means the right side of the expression
		// is finished
		if contains, op := hasSuffixComparisonOp(tmpString); contains {
			if foundOp { // if there was already a found operator for this simple expression, return error
				return MetricFilterPattern{}, errors.New("got multiple comparison operators")
			}

			// Remove the operator suffix and trailing spaces
			left = strings.TrimSpace(strings.TrimSuffix(tmpString, string(op)))
			operator = op
			foundOp = true
			buf.Reset()
			buf.Grow(len(s) - i)
		}
	}

	if !foundOp {
		return MetricFilterPattern{}, errors.New("could not find a ComparisonOperator for this MetricFilterPattern")
	}

	// Trim trailing spaces
	right := strings.TrimSpace(buf.String())
	return newSimpleExpression(left, operator, right), nil
}

func hasSuffixComparisonOp(s string) (bool, comparisonOperator) {
	for _, op := range listComparisonOperators() {
		if strings.HasSuffix(s, string(op)) {
			return true, op
		}
	}
	return false, ""
}

func hasSuffixLogicalOp(s string) (bool, logicalOperator) {
	for _, op := range listLogicalOperators() {
		if strings.HasSuffix(s, string(op)) {
			return true, op
		}
	}
	return false, ""
}
