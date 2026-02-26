// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobra

import (
	"strings"
	"testing"
)

func TestNoBoolFlagCreated(t *testing.T) {
	var dryRun bool

	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", true, "Show actions without executing")

	output, err := executeCommand(rootCmd, "--no-dry-run")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if output != "" {
		t.Fatalf("Unexpected output: %s", output)
	}
	if dryRun != false {
		t.Errorf("expected dryRun to be false, got true")
	}
}

func TestNoBoolFlagDefaultBehavior(t *testing.T) {
	var dryRun bool

	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", true, "Show actions without executing")

	// Without any flags, dryRun should stay true (default)
	_, err := executeCommand(rootCmd, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dryRun != true {
		t.Errorf("expected dryRun to be true (default), got false")
	}
}

func TestNoBoolFlagMutuallyExclusive(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().Bool("dry-run", true, "Show actions without executing")

	_, err := executeCommand(rootCmd, "--dry-run", "--no-dry-run")
	if err == nil {
		t.Error("expected error when both --dry-run and --no-dry-run are set, got nil")
	}
}

func TestNoBoolFlagNotCreatedForFalseDefault(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().Bool("verbose", false, "verbose output")

	// initNoBoolFlags runs during execute, so trigger it
	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if rootCmd.Flags().Lookup("no-verbose") != nil {
		t.Error("expected no --no-verbose flag for a flag defaulting to false")
	}
}

func TestNoBoolFlagNotCreatedWhenDisabled(t *testing.T) {
	rootCmd := &Command{
		Use: "root",
		Run: emptyRun,
		// AddNegativeBoolFlags is false (default)
	}
	rootCmd.Flags().Bool("dry-run", true, "Show actions without executing")

	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if rootCmd.Flags().Lookup("no-dry-run") != nil {
		t.Error("expected no --no-dry-run flag when AddNegativeBoolFlags is false")
	}
}

func TestNoBoolFlagSkipsExistingNoPrefix(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().Bool("no-cache", true, "disable cache")

	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not create --no-no-cache
	if rootCmd.Flags().Lookup("no-no-cache") != nil {
		t.Error("expected no --no-no-cache flag for a flag already starting with no-")
	}
}

func TestNoBoolFlagSkipsExistingManualFlag(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().Bool("dry-run", true, "Show actions without executing")
	rootCmd.Flags().Bool("no-dry-run", false, "manually defined negation")

	// Should not panic or error - the manual --no-dry-run takes precedence
	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestNoBoolFlagWithPersistentFlags(t *testing.T) {
	var dryRun bool

	rootCmd := &Command{
		Use:                  "root",
		AddNegativeBoolFlags: true,
	}
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", true, "Show actions without executing")

	childCmd := &Command{
		Use: "child",
		Run: emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	// --no-dry-run should work on the child command via inheritance
	_, err := executeCommand(rootCmd, "child", "--no-dry-run")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dryRun != false {
		t.Errorf("expected dryRun to be false, got true")
	}
}

func TestNoBoolFlagInheritedFromParent(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		AddNegativeBoolFlags: true,
	}

	childCmd := &Command{
		Use: "child",
		Run: emptyRun,
	}
	childCmd.Flags().Bool("dry-run", true, "Show actions without executing")
	rootCmd.AddCommand(childCmd)

	// AddNegativeBoolFlags on root should be inherited by child
	_, err := executeCommand(rootCmd, "child", "--no-dry-run")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestNoBoolFlagCompletion(t *testing.T) {
	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().Bool("dry-run", true, "Show actions without executing")
	rootCmd.Flags().String("output", "", "output file")

	// Completing --no- should suggest --no-dry-run
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--no-")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"--no-dry-run",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Completing --no (without trailing dash) should also suggest --no-dry-run
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--no")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestNoBoolFlagWithExplicitValue(t *testing.T) {
	var dryRun bool

	rootCmd := &Command{
		Use:                  "root",
		Run:                  emptyRun,
		AddNegativeBoolFlags: true,
	}
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", true, "Show actions without executing")

	// --no-dry-run=false should set dry-run back to true (double negation)
	_, err := executeCommand(rootCmd, "--no-dry-run=false")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dryRun != true {
		t.Errorf("expected dryRun to be true (double negation), got false")
	}
}
