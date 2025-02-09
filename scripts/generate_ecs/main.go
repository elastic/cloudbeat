package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "github.com/elastic/beats/v7/libbeat/ecs"
	"golang.org/x/tools/go/packages"
)

var files = []string{
	"cloud.go",
	"container.go",
	"process.go",
	"vulnerability.go",
	"base.go",
	"event.go",
	"network.go",
	"host.go",
	"orchestration.go",
	"organization.go",
	"related.go",
	"user.go",
	"service.go",
}

func replaceEcsWithJsonTags(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filePath, content, parser.ParseComments) // Important: Parse with comments
	if err != nil {
		return err
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if field.Tag != nil {
					field.Tag.Value = fmt.Sprintf("`json:\"%s,omitempty\"`", strings.Trim(field.Tag.Value[5:len(field.Tag.Value)-1], `"`))
				}
			}
		}
	}

	var buf bytes.Buffer
	err = printer.Fprint(&buf, fs, node)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

func processPackage(pkgPath, outputDir string) error {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return fmt.Errorf("Error loading package: %v", err)
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.CompiledGoFiles {
			fileName := filepath.Base(file)
			if !slices.Contains(files, fileName) {
				continue
			}

			fmt.Printf("Processing file: %s\n", fileName)
			targetFilePath := filepath.Join(outputDir, fileName)
			err := os.MkdirAll(filepath.Dir(targetFilePath), 0755)
			if err != nil {
				return fmt.Errorf("Error creating target directory: %v", err)
			}

			sourceFileContent, err := os.ReadFile(file) // Read the file content from source path
			if err != nil {
				return fmt.Errorf("Error reading source file: %v", err)
			}

			err = os.WriteFile(targetFilePath, sourceFileContent, 0644)
			if err != nil {
				return fmt.Errorf("Error copying file to target directory: %v", err)
			}

			err = replaceEcsWithJsonTags(targetFilePath)
			if err != nil {
				return fmt.Errorf("Error replacing ecs tags: %v", err)
			}
		}
	}

	return nil
}

func main() {
	const ecsPackage = "github.com/elastic/beats/v7/libbeat/ecs"
	outputDir := "../../internal/ecs"
	err := processPackage(ecsPackage, outputDir)
	if err != nil {
		log.Fatalf("Error processing package: %v\n", err)
	}

	log.Println("ECS files have been successfully generated and updated.")
}
