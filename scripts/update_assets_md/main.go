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

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/ettle/strcase"
	"github.com/xuri/excelize/v2"
)

const (
	GO_FILE             = "../../internal/inventory/asset.go"
	EXCEL_FILE          = "../../internal/inventory/cloud_assets.xlsx"
	SUMMARY_FILE        = "../../internal/inventory/ASSETS.md"
	CLASSIFICATION_TYPE = "AssetClassification"
	// Provider prefixes
	AWS_PREFIX   = "Aws"
	AZURE_PREFIX = "Azure"
	GCP_PREFIX   = "Gcp"
)

type Classification struct {
	Category    string
	SubCategory string
	Type        string
	SubType     string
}

func (item Classification) ID() string {
	return fmt.Sprintf("%s_%s_%s_%s",
		strcase.ToKebab(item.Category),
		strcase.ToKebab(item.SubCategory),
		strcase.ToKebab(item.Type),
		strcase.ToKebab(item.SubType),
	)
}

type ByProvider struct {
	AWS   map[string]Classification
	Azure map[string]Classification
	GCP   map[string]Classification
}

func (bp *ByProvider) Assign(provider string, c Classification) {
	switch provider {
	case AWS_PREFIX:
		bp.AWS[c.ID()] = c
	case AZURE_PREFIX:
		bp.Azure[c.ID()] = c
	case GCP_PREFIX:
		bp.GCP[c.ID()] = c
	default:
		panic(fmt.Errorf("unsupported provider: %s", provider))
	}
}

func (bp *ByProvider) Get(provider string) map[string]Classification {
	switch provider {
	case AWS_PREFIX:
		return bp.AWS
	case AZURE_PREFIX:
		return bp.Azure
	case GCP_PREFIX:
		return bp.GCP
	default:
		panic(fmt.Errorf("unsupported provider: %s", provider))
	}
}

func main() {
	implementedByProvider, err := loadClassificationsFromGolang(GO_FILE)
	if err != nil {
		panic(err)
	}
	plannedByProvider, err := loadClassificationsFromExcel(EXCEL_FILE)
	if err != nil {
		panic(err)
	}
	err = writeSummary(plannedByProvider, implementedByProvider, SUMMARY_FILE)
	if err != nil {
		panic(err)
	}
}

func loadClassificationsFromGolang(filepath string) (*ByProvider, error) {
	output := &ByProvider{
		AWS:   map[string]Classification{},
		Azure: map[string]Classification{},
		GCP:   map[string]Classification{},
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return output, fmt.Errorf("failed to parse Go file: %w", err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		for _, spec := range decl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			provider, ok := extractProvider(valSpec)
			if !ok {
				continue
			}

			for _, value := range valSpec.Values {
				cl, err := extractClassification(value)
				if err != nil {
					continue
				}
				output.Assign(provider, cl)
			}
		}

		return false
	})

	return output, nil
}

func loadClassificationsFromExcel(filepath string) (*ByProvider, error) {
	output := &ByProvider{
		AWS:   map[string]Classification{},
		Azure: map[string]Classification{},
		GCP:   map[string]Classification{},
	}

	f, err := excelize.OpenFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	sheets := map[int]string{1: AWS_PREFIX, 2: AZURE_PREFIX, 3: GCP_PREFIX}
	for sheetNo, provider := range sheets {
		sheetName := f.GetSheetName(sheetNo)
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to get rows: %w", err)
		}

		headers := rows[0]
		for _, row := range rows[1:] {
			cl := Classification{
				Category:    row[getColumnIndex(headers, "asset.category")],
				SubCategory: row[getColumnIndex(headers, "asset.subcategory")],
				Type:        row[getColumnIndex(headers, "asset.type")],
				SubType:     row[getColumnIndex(headers, "asset.subtype")],
			}
			output.Assign(provider, cl)
		}
	}

	return output, nil
}

func writeSummary(plannedByProvider, implementedByProvider *ByProvider, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for providerNo, provider := range []string{AWS_PREFIX, AZURE_PREFIX, GCP_PREFIX} {
		planned := plannedByProvider.Get(provider)
		implemented := implementedByProvider.Get(provider)
		sortedKeys := slices.Sorted(maps.Keys(planned))

		// stats
		totalImplemented := 0
		implementedByCategory := map[string]int{}
		plannedByCategory := map[string]int{}

		// table of assets
		table := []string{
			"<details> <summary>Full table</summary>\n",
			"| Category | SubCategory | Type | SubType | Implemented? |",
			"|---|---|---|---|---|",
		}

		for _, key := range sortedKeys {
			item := planned[key]
			status := "No ❌"

			plannedByCategory[item.Category] += 1
			if _, ok := implemented[key]; ok {
				status = "Yes ✅"
				totalImplemented += 1
				implementedByCategory[item.Category] += 1
			}
			table = append(table,
				fmt.Sprintf(
					"| %s | %s | %s | %s | %s |",
					item.Category, item.SubCategory, item.Type, item.SubType, status,
				),
			)
		}
		table = append(table, "\n</details>")

		// write ASSETS.md
		if providerNo > 0 {
			writeToFile(file, "\n\n")
		}
		writeToFile(file, fmt.Sprintf("## %s Resources\n\n", strings.ToUpper(provider)))

		percentage := totalImplemented * 100 / len(planned)
		writeToFile(
			file,
			fmt.Sprintf("**Progress: %d%% (%d/%d)**\n", percentage, totalImplemented, len(planned)),
		)

		sortedCategories := maps.Keys(plannedByCategory)
		slices.Sort(sortedCategories)

		for _, category := range sortedCategories {
			plannedCount := plannedByCategory[category]
			implementedCount := implementedByCategory[category]
			percentage = implementedCount * 100 / plannedCount
			writeToFile(
				file,
				fmt.Sprintf("%s: %d%% (%d/%d)\n", category, percentage, implementedCount, plannedCount),
			)
		}
		writeToFile(file, "\n"+strings.Join(table, "\n"))
	}
	writeToFile(file, "\n") // required for valid Markdown :o

	return nil
}

// Golang AST functions -------------------------------------------------

func extractProvider(valSpec *ast.ValueSpec) (string, bool) {
	if len(valSpec.Names) == 0 {
		return "", false
	}
	name := valSpec.Names[0].Name
	if !strings.HasPrefix(name, CLASSIFICATION_TYPE) {
		return "", false
	}
	name = name[len(CLASSIFICATION_TYPE):]

	if strings.HasPrefix(name, AWS_PREFIX) {
		return AWS_PREFIX, true
	}
	if strings.HasPrefix(name, AZURE_PREFIX) {
		return AZURE_PREFIX, true
	}
	if strings.HasPrefix(name, GCP_PREFIX) {
		return GCP_PREFIX, true
	}
	return "", false
}

func extractClassification(expr ast.Expr) (Classification, error) {
	output := Classification{}

	compLit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return output, fmt.Errorf("value is not a composite literal, skipping")
	}

	if len(compLit.Elts) != 4 {
		return output, fmt.Errorf("expected full, 4-field classification; got %d", len(compLit.Elts))
	}

	classification := []string{}
	for _, elt := range compLit.Elts {
		s, err := unpackElt(elt)
		if err != nil {
			return output, fmt.Errorf("could not unpack elements: %w", err)
		}
		classification = append(classification, s)
	}

	output.Category = classification[0]
	output.SubCategory = classification[1]
	output.Type = classification[2]
	output.SubType = classification[3]
	return output, nil
}

func unpackElt(expr ast.Expr) (string, error) {
	switch element := expr.(type) {
	case *ast.KeyValueExpr:
		return unpackKeyValueExpr(element)
	case *ast.BasicLit:
		return unpackBasicLit(element)
	case *ast.Ident:
		return unpackIdent(element)
	default:
		return "", fmt.Errorf("unhandled expr type %T", element)
	}
}

func unpackIdent(obj *ast.Ident) (string, error) {
	o := obj.Obj
	if o == nil {
		return "", nil
	}
	valueSpec, ok := o.Decl.(*ast.ValueSpec)
	if !ok {
		return "", fmt.Errorf("cannot cast values to []Expr")
	}
	if len(valueSpec.Values) != 1 {
		return "", fmt.Errorf("this should not happen - len(values) != 1")
	}
	basicLitVal, ok := valueSpec.Values[0].(*ast.BasicLit)
	if !ok {
		return "", fmt.Errorf("got a single value, but it's not a BasicLit")
	}
	return unpackBasicLit(basicLitVal)
}

func unpackBasicLit(obj *ast.BasicLit) (string, error) {
	return strings.Trim(obj.Value, "\" "), nil
}

func unpackKeyValueExpr(obj *ast.KeyValueExpr) (string, error) {
	switch v := obj.Value.(type) {
	case *ast.Ident:
		return unpackIdent(v)
	case *ast.BasicLit:
		return unpackBasicLit(v)
	default:
		return "", fmt.Errorf("cannot unpack KeyValue val")
	}
}

// Helper functions -----------------------------------------------------

// Helper function to get the index of a header in the Excel file
func getColumnIndex(headers []string, headerName string) int {
	for i, h := range headers {
		if strings.TrimSpace(h) == headerName {
			return i
		}
	}
	return -1
}

func writeToFile(file *os.File, s string) {
	_, err := file.WriteString(s)
	if err != nil {
		panic(err)
	}
}
