package planner_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
)

func TestExecuteWritesGeneratedFilesAndEnv(t *testing.T) {
	service := planner.New(registry.New())
	plan := core.TemplatePlan{
		Actions: []core.ExecutionAction{
			{
				ID:          "write-file",
				Name:        "Write app config",
				Description: "Creates a generated config file.",
				Kind:        core.ActionKindWriteFile,
				Path:        filepath.Join("demo", "config", "app.json"),
				Content:     "{\n  \"name\": \"demo\"\n}\n",
			},
			{
				ID:          "write-env",
				Name:        "Write env template",
				Description: "Creates the env example file.",
				Kind:        core.ActionKindWriteEnv,
				Path:        filepath.Join("demo", ".env.example"),
				Content:     "DATABASE_URL=postgres://localhost:5432/demo\n",
			},
		},
	}

	workDir := t.TempDir()
	restoreWorkingDir := chdirForTest(t, workDir)
	defer restoreWorkingDir()

	if err := service.Execute(context.Background(), plan, io.Discard, io.Discard); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	configFilePath := filepath.Join(workDir, "demo", "config", "app.json")
	configFileContent := readFile(t, configFilePath)
	if configFileContent != "{\n  \"name\": \"demo\"\n}\n" {
		t.Fatalf("unexpected config file content: %q", configFileContent)
	}

	envFilePath := filepath.Join(workDir, "demo", ".env.example")
	envFileContent := readFile(t, envFilePath)
	if envFileContent != "DATABASE_URL=postgres://localhost:5432/demo\n" {
		t.Fatalf("unexpected env file content: %q", envFileContent)
	}
}

func TestFormatPrintsResolvedTemplateSummaryOnly(t *testing.T) {
	service := planner.New(registry.New())
	plan := core.TemplatePlan{
		Selection: func() core.TemplateSelection {
			selection := core.NewTemplateSelection()
			selection.ProjectName = "demo"
			selection.SetSingle(core.GroupFrontend, "next")
			selection.SetSingle(core.GroupBackend, "hono")
			selection.SetSingle(core.GroupDatabase, "supabase")
			selection.SetMulti(core.GroupStorage, []string{"r2-storage"})
			selection.SetSingle(core.GroupORM, "drizzle")
			selection.SetSingle(core.GroupPackageManager, "pnpm")
			selection.SetSingle(core.GroupCodeQuality, "biome")
			selection.SetMulti(core.GroupAddon, []string{"better-auth"})
			return selection
		}(),
		Fragments: []core.PlanFragment{
			{
				ID:      "generated-files",
				OwnerID: "test-owner",
				Phase:   core.PhaseIntegration,
				Actions: []core.ExecutionAction{
					{
						ID:          "write-file",
						Name:        "Write file",
						Description: "Creates a generated file.",
						Kind:        core.ActionKindWriteFile,
						Path:        filepath.Join("demo", "file.txt"),
					},
					{
						ID:          "write-env",
						Name:        "Write env",
						Description: "Creates an env file.",
						Kind:        core.ActionKindWriteEnv,
						Path:        filepath.Join("demo", ".env.example"),
					},
				},
			},
		},
	}

	output := service.Format(plan)
	for _, expected := range []string{"Resolved template", "Project          demo", "Backend          hono", "Storage          r2-storage", "ORM              drizzle", "Addons           better-auth"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected formatted summary to contain %q, got:\n%s", expected, output)
		}
	}
	for _, unexpected := range []string{"Plan fragments", "write file demo/file.txt", "write env file demo/.env.example"} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("expected formatted summary to omit %q, got:\n%s", unexpected, output)
		}
	}
}

func chdirForTest(t *testing.T, target string) func() {
	t.Helper()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := os.Chdir(target); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	return func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatalf("restore Chdir() error = %v", err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(content)
}
