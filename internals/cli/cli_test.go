package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
	"openblueprints/internals/resolver"
	"openblueprints/internals/tui"
)

func TestParseArgsDoesNotExposePackageInstallFinalization(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	finalization := selection.Multi(core.GroupFinalization)
	if len(finalization) != 0 {
		t.Fatalf("expected no package-install finalization choice, got %v", finalization)
	}
}

func TestParseArgsAllowsDisablingFinalization(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo", "--finalization="}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	if !selection.IsTouched(core.GroupFinalization) {
		t.Fatalf("expected finalization to be touched when disabled explicitly")
	}
	if values := selection.Multi(core.GroupFinalization); len(values) != 0 {
		t.Fatalf("expected empty finalization, got %v", values)
	}
}

func TestParseArgsIgnoresRemovedInstallPackagesFinalization(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo", "--finalization", "install-packages,git-init"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	finalization := selection.Multi(core.GroupFinalization)
	if len(finalization) != 1 || finalization[0] != "git-init" {
		t.Fatalf("expected only git-init finalization, got %v", finalization)
	}
}

func TestParseArgsInstallSkillsFlagAddsFinalization(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo", "--install-skills"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	finalization := selection.Multi(core.GroupFinalization)
	expected := []string{"install-agent-skills"}
	if len(finalization) != len(expected) {
		t.Fatalf("expected finalization %v, got %v", expected, finalization)
	}
	for index, expectedValue := range expected {
		if finalization[index] != expectedValue {
			t.Fatalf("expected finalization %v, got %v", expected, finalization)
		}
	}
}

func TestParseArgsSplitsStorageOutOfLegacyAddonsFlag(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo", "--addons", "better-auth,r2-storage"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	addons := selection.Multi(core.GroupAddon)
	if len(addons) != 1 || addons[0] != "better-auth" {
		t.Fatalf("expected addons to contain only better-auth, got %v", addons)
	}
	storage := selection.Multi(core.GroupStorage)
	if len(storage) != 1 || storage[0] != "r2-storage" {
		t.Fatalf("expected storage to contain r2-storage, got %v", storage)
	}
}

func TestParseArgsAcceptsCodeQualityAndStorageFlags(t *testing.T) {
	selection, _, _, _, err := parseArgs([]string{"--name", "demo", "--code-quality", "oxlint-oxformat", "--storage", "s3-storage"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}

	if selection.Single(core.GroupCodeQuality) != "oxlint-oxformat" {
		t.Fatalf("expected oxlint-oxformat code quality, got %q", selection.Single(core.GroupCodeQuality))
	}
	storage := selection.Multi(core.GroupStorage)
	if len(storage) != 1 || storage[0] != "s3-storage" {
		t.Fatalf("expected storage to contain s3-storage, got %v", storage)
	}
}

func TestRunCompletesDefaultCLIStack(t *testing.T) {
	app := testApp()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := app.Run([]string{"--name", "demo", "--preview"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	for _, expected := range []string{"Frontend         next", "Backend          express", "Database         postgres", "ORM              prisma", "Linting / Formatter biome"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "install-packages") {
		t.Fatalf("did not expect install-packages in output, got:\n%s", output)
	}
}

func TestRunCompletesMongoCLIStackWithMongoose(t *testing.T) {
	app := testApp()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := app.Run([]string{"--name", "demo", "--database", "mongodb", "--preview"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v, stderr = %s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "ORM              mongoose") {
		t.Fatalf("expected CLI defaults to select mongoose for MongoDB, got:\n%s", stdout.String())
	}
}

func testApp() App {
	reg := registry.New()
	builder := planner.New(reg)
	planResolver := resolver.New(reg, builder)
	return New(planResolver, builder, tui.New(planResolver, builder))
}
