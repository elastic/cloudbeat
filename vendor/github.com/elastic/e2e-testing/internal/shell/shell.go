// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.elastic.co/apm"
)

// CheckInstalledSoftware checks that the required software is present
func CheckInstalledSoftware(binaries ...string) {
	log.Tracef("Validating required tools: %v", binaries)

	for _, binary := range binaries {
		err := which(binary)
		if err != nil {
			log.Fatalf("The program cannot be run because %s are not installed. Required: %v", binary, binaries)
		}
	}
}

// Execute executes a command in the machine the program is running
// - workspace: represents the location where to execute the command
// - command: represents the name of the binary to execute
// - args: represents the arguments to be passed to the command
func Execute(ctx context.Context, workspace string, command string, args ...string) (string, error) {
	return ExecuteWithEnv(ctx, workspace, command, map[string]string{}, args...)
}

// ExecuteWithEnv executes a command in the machine the program is running
// - workspace: represents the location where to execute the command
// - command: represents the name of the binary to execute
// - env: represents the environment variables to be passed to the command
// - args: represents the arguments to be passed to the command
func ExecuteWithEnv(ctx context.Context, workspace string, command string, env map[string]string, args ...string) (string, error) {
	return ExecuteWithStdin(ctx, workspace, nil, command, env, args...)
}

// ExecuteWithStdin executes a command in the machine the program is running
// - workspace: represents the location where to execute the command
// - stdin: reader to use as standard input for the command
// - command: represents the name of the binary to execute
// - args: represents the arguments to be passed to the command
func ExecuteWithStdin(ctx context.Context, workspace string, stdin io.Reader, command string, env map[string]string, args ...string) (string, error) {
	span, _ := apm.StartSpanOptions(ctx, "Executing shell command", "shell.command.execute", apm.SpanOptions{
		Parent: apm.SpanFromContext(ctx).TraceContext(),
	})
	span.Context.SetLabel("workspace", workspace)
	span.Context.SetLabel("command", command)
	span.Context.SetLabel("arguments", args)
	span.Context.SetLabel("environment", env)
	defer span.End()

	log.WithFields(log.Fields{
		"command": command,
		"args":    args,
		"env":     env,
	}).Trace("Executing command")

	cmd := exec.Command(command, args[0:]...)

	cmd.Dir = workspace

	if len(env) > 0 {
		environment := os.Environ()

		for k, v := range env {
			environment = append(environment, fmt.Sprintf("%s=%s", k, v))
		}

		cmd.Env = environment
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if stdin != nil {
		cmd.Stdin = stdin
	}

	err := cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"baseDir": workspace,
			"command": command,
			"args":    args,
			"env":     env,
			"error":   err,
			"stderr":  stderr.String(),
		}).Error("Error executing command")

		return "", err
	}

	trimmedOutput := strings.Trim(out.String(), "\n")

	log.WithFields(log.Fields{
		"output": trimmedOutput,
	}).Trace("Output")

	return trimmedOutput, nil
}

// GetEnv returns an environment variable as string
func GetEnv(envVar string, defaultValue string) string {
	value, exists := os.LookupEnv(envVar)
	if exists && value != "" {
		return value
	}

	return defaultValue
}

// GetEnvBool returns an environment variable as boolean.
// If the variable is not present, returns false
func GetEnvBool(key string) bool {
	s := os.Getenv(key)
	if s == "" {
		return false
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}

	return v
}

// GetEnvInteger returns an environment variable as integer, including a default value
func GetEnvInteger(envVar string, defaultValue int) int {
	if value, exists := os.LookupEnv(envVar); exists {
		v, err := strconv.Atoi(value)
		if err == nil {
			return v
		}
	}

	return defaultValue
}

// which checks if software is installed, else it aborts the execution
func which(binary string) error {
	path, err := exec.LookPath(binary)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"binary": binary,
		}).Error("Required binary is not present")
		return err
	}

	log.WithFields(log.Fields{
		"binary": binary,
		"path":   path,
	}).Trace("Binary is present")
	return nil
}
