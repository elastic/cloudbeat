//go:build mage
// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"

	devtools "github.com/elastic/beats/v7/dev-tools/mage"
	cloudbeat "github.com/elastic/cloudbeat/scripts/mage"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/pkg"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/unittest"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/integtest/notests"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/test"
)

func init() {
	repo, err := devtools.GetProjectRepoInfo()
	if err != nil {
		panic(err)
	}

	devtools.BeatDescription = "Cloudbeat collects cloud compliance and sends findings to ElasticSearch"
	devtools.BeatLicense = "Elastic License"
	devtools.SetBuildVariableSources(&devtools.BuildVariableSources{
		BeatVersion: filepath.Join(repo.RootDir, "version.go"),
		GoVersion:   filepath.Join(repo.RootDir, ".go-version"),
		DocBranch:   filepath.Join(repo.RootDir, "docs/version.asciidoc"),
	})
}

// Check formats code, updates generated content, check for common errors, and
// checks for any modified files.
func Check() error {
	return devtools.Check()
}

// Build builds the Beat binary.
func Build() error {
	return devtools.Build(devtools.DefaultBuildArgs())
}

// Clean cleans all generated files and build artifacts.
func Clean() error {
	return devtools.Clean()
}

// Update updates the generated files (aka make update).

// GolangCrossBuild build the Beat binary inside of the golang-builder.
// Do not use directly, use crossBuild instead.
func GolangCrossBuild() error {
	return devtools.GolangCrossBuild(devtools.DefaultGolangCrossBuildArgs())
}

// BuildGoDaemon builds the go-daemon binary (use crossBuildGoDaemon).
func BuildGoDaemon() error {
	return devtools.BuildGoDaemon()
}

// CrossBuild cross-builds the beat for all target platforms.
func CrossBuild() error {
	return devtools.CrossBuild()
}

// CrossBuildGoDaemon cross-builds the go-daemon binary using Docker.
func CrossBuildGoDaemon() error {
	return devtools.CrossBuildGoDaemon()
}

// Package packages the Beat for distribution.
// Use SNAPSHOT=true to build snapshots.
// Use PLATFORMS to control the target platforms.
// Use VERSION_QUALIFIER to control the version qualifier.
func Package() {
	start := time.Now()
	defer func() { fmt.Println("package ran for", time.Since(start)) }()

	//devtools.MustUsePackaging("cloudbeat", "cloudbeat/dev-tools/packaging/packages.yml")

	devtools.UseElasticBeatOSSPackaging()
	cloudbeat.CustomizePackaging()

	if packageTypes := os.Getenv("TYPES"); packageTypes != "" {
		filterPackages(packageTypes)
	}

	mg.Deps(Update)
	mg.Deps(CrossBuild, CrossBuildGoDaemon)
	mg.SerialDeps(devtools.Package, TestPackages)
}

func keepPackages(types []string) map[devtools.PackageType]struct{} {
	keep := make(map[devtools.PackageType]struct{})
	for _, t := range types {
		var pt devtools.PackageType
		if err := pt.UnmarshalText([]byte(t)); err != nil {
			log.Printf("skipped filtering package type %s", t)
			continue
		}
		keep[pt] = struct{}{}
	}
	return keep
}

func filterPackages(types string) {
	var packages []devtools.OSPackageArgs
	keep := keepPackages(strings.Split(types, " "))
	for _, p := range devtools.Packages {
		for _, t := range p.Types {
			if _, ok := keep[t]; !ok {
				continue
			}
			packages = append(packages, p)
			break
		}
	}
	devtools.Packages = packages
}

// TestPackages tests the generated packages (i.e. file modes, owners, groups).
func TestPackages() error {
	return devtools.TestPackages()
}

func Update() { mg.Deps(cloudbeat.Update.All) }

// Fields generates a fields.yml for the Beat.
func Fields() { mg.Deps(cloudbeat.Update.Fields) }

// Config generates both the short/reference/docker configs.
func Config() { mg.Deps(cloudbeat.Update.Config) }
