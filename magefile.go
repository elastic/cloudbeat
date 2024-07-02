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

//go:build mage

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/elastic/beats/v7/dev-tools/mage"
	devtools "github.com/elastic/beats/v7/dev-tools/mage"
	"github.com/elastic/beats/v7/dev-tools/mage/gotool"
	"github.com/elastic/e2e-testing/pkg/downloads"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/integtest/notests"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/pkg"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/test"
	// mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/unittest"

	cloudbeat "github.com/elastic/cloudbeat/scripts/mage"
)

const (
	snapshotEnv   = "SNAPSHOT"
	agentDropPath = "AGENT_DROP_PATH"
)

func init() {
	repo, err := devtools.GetProjectRepoInfo()
	if err != nil {
		panic(err)
	}

	devtools.BeatDescription = "Cloudbeat collects cloud compliance data and sends findings to ElasticSearch"
	devtools.BeatLicense = "Elastic License"
	devtools.SetBuildVariableSources(&devtools.BuildVariableSources{
		BeatVersion: filepath.Join(repo.RootDir, "version/version.go"),
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
	mg.Deps(BuildOpaBundle)

	args := devtools.DefaultBuildArgs()
	args.CGO = false
	return devtools.Build(args)
}

// Clean cleans all generated files and build artifacts.
func Clean() error {
	return devtools.Clean()
}

// Update updates the generated files (aka make update).

// GolangCrossBuild build the Beat binary inside of the golang-builder.
// Do not use directly, use crossBuild instead.
func GolangCrossBuild() error {
	args := devtools.DefaultGolangCrossBuildArgs()
	args.CGO = false
	return devtools.GolangCrossBuild(args)
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

// Run UnitTests
func GoTestUnit(ctx context.Context) error {
	return devtools.GoTest(ctx, devtools.DefaultGoTestUnitArgs())
}

// Package packages the Beat for distribution.
// Use SNAPSHOT=true to build snapshots.
// Use PLATFORMS to control the target platforms.
// Use VERSION_QUALIFIER to control the version qualifier.
func Package() {
	start := time.Now()
	defer func() { fmt.Println("package ran for", time.Since(start)) }()

	devtools.UseElasticBeatXPackPackaging()
	cloudbeat.CustomizePackaging()

	if packageTypes := os.Getenv("TYPES"); packageTypes != "" {
		filterPackages(packageTypes)
	}

	mg.Deps(Update)
	mg.Deps(CrossBuild, CrossBuildGoDaemon)
	mg.SerialDeps(devtools.Package)
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

// Fmt formats code and adds license headers.
func Fmt() {
	mg.Deps(devtools.GoImports, devtools.PythonAutopep8)
	mg.Deps(AddLicenseHeaders)
}

// AddLicenseHeaders adds ASL2 headers to .go files outside of x-pack and
// add Elastic headers to .go files in x-pack.
func AddLicenseHeaders() error {
	fmt.Println(">> fmt - go-licenser: Adding missing headers")

	mg.Deps(devtools.InstallGoLicenser)

	licenser := gotool.Licenser

	return licenser(
		licenser.License("ASL2"),
		licenser.Exclude("x-pack"),
	)
}

// CheckLicenseHeaders checks ASL2 headers in .go files outside of x-pack and
// checks Elastic headers in .go files in x-pack.
func CheckLicenseHeaders() error {
	fmt.Println(">> fmt - go-licenser: Checking for missing headers")

	mg.Deps(devtools.InstallGoLicenser)

	licenser := gotool.Licenser

	return licenser(
		licenser.Check(),
		licenser.License("ASL2"),
	)
}

func Update() {
	mg.Deps(cloudbeat.Update.All, BuildOpaBundle)
	mg.Deps(AddLicenseHeaders)
}

func bundleAgent() {
	pwd, err := filepath.Abs("../elastic-agent")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("mage", "dev:package")
	cmd.Dir = pwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("PWD=%s", pwd), "TYPES=docker")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func PackageAgent() {
	beatVersion, found := os.LookupEnv("BEAT_VERSION")
	if !found {
		beatVersion, _ = devtools.BeatQualifiedVersion()
	}
	// prepare new drop
	dropPath := filepath.Join("build", "elastic-agent-drop")
	dropPath, err := filepath.Abs(dropPath)
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dropPath, 0o755); err != nil {
		panic(err)
	}

	os.Setenv(agentDropPath, dropPath)

	// cleanup after build
	defer os.RemoveAll(dropPath)
	defer os.Unsetenv(agentDropPath)

	platformPackages := []struct {
		platform string
		packages string
	}{
		{"darwin/amd64", "darwin-x86_64.tar.gz"},
		{"darwin/arm64", "darwin-aarch64.tar.gz"},
		{"linux/amd64", "linux-x86_64.tar.gz"},
		{"linux/arm64", "linux-arm64.tar.gz"},
		{"windows/amd64", "windows-x86_64.zip"},
	}

	var requiredPackages []string
	for _, p := range platformPackages {
		if _, enabled := devtools.Platforms.Get(p.platform); enabled {
			requiredPackages = append(requiredPackages, p.packages)
		}
	}

	packedBeats := []string{"filebeat", "heartbeat", "metricbeat", "osquerybeat"}
	ctx := context.Background()
	for _, beat := range packedBeats {
		for _, reqPackage := range requiredPackages {
			newVersion, packageName := getPackageName(beat, beatVersion, reqPackage)
			err := fetchBinaryFromArtifactsApi(ctx, packageName, beat, newVersion, dropPath)
			if err != nil {
				panic(fmt.Sprintf("fetchBinaryFromArtifactsApi failed: %v", err))
			}
		}
	}
	mg.Deps(Package)

	// copy to new drop
	sourcePath := filepath.Join("build", "distributions")
	if err := copyAll(sourcePath, dropPath); err != nil {
		panic(err)
	}
	mg.Deps(bundleAgent)
}

func getPackageName(beat, version, pkg string) (string, string) {
	if _, ok := os.LookupEnv(snapshotEnv); ok {
		version += "-SNAPSHOT"
	}
	return version, fmt.Sprintf("%s-%s-%s", beat, version, pkg)
}

func fetchBinaryFromArtifactsApi(ctx context.Context, packageName, artifact, version, downloadPath string) error {
	location, err := downloads.FetchBeatsBinary(
		ctx,
		packageName,
		artifact,
		version,
		3,
		false,
		downloadPath,
		true)
	fmt.Println("downloaded binaries on location:", location)
	return err
}

func copyAll(from, to string) error {
	return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		targetFile := filepath.Join(to, info.Name())

		// overwrites with current build
		return sh.Copy(targetFile, path)
	})
}

// Fields generates a fields.yml for the Beat.
func Fields() { mg.Deps(cloudbeat.Update.Fields) }

// Config generates both the short/reference/docker configs.
func Config() { mg.Deps(cloudbeat.Update.Config) }

// PythonEnv ensures the Python venv is up-to-date with the beats requirements.txt.
func PythonEnv() error {
	_, err := mage.PythonVirtualenv(true)
	return err
}

func getMajorMinorVersion(version string) string {
	return strings.Join(strings.Split(version, ".")[:2], ".")
}

func BuildOpaBundle() error {
	return sh.Run("bin/opa", "build", "-b", "security-policies/bundle", "-e", "security-policies/bundle/compliance")
}
