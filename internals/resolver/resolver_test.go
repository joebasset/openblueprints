package resolver_test

import (
	"strings"
	"testing"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
	"openblueprints/internals/resolver"
)

func TestResolveNextMongoExcludesPostgresORMs(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "mongodb")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupORM {
		t.Fatalf("expected orm group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"mongoose"})
}

func TestResolveNextPostgresIncludesExpectedORMs(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupORM {
		t.Fatalf("expected orm group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"drizzle", "prisma"})
}

func TestResolveNextSupabaseIncludesSQLORMs(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "supabase")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupORM {
		t.Fatalf("expected orm group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"drizzle", "prisma"})
}

func TestResolveNextGoSkipsORM(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "go-api")
	selection.SetSingle(core.GroupDatabase, "postgres")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupPackageManager {
		t.Fatalf("expected package manager group, got %#v", group)
	}
}

func TestResolveNextUnsupportedAddonHidden(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "go-api")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupPackageManager, "npm")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group != nil {
		t.Fatalf("expected no addon group for unsupported stack, got %#v", group)
	}
}

func TestResolveNextAddonShownForPrismaPath(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth"})
}

func TestResolveNextAddonShownForMongoPath(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "mongodb")
	selection.SetSingle(core.GroupORM, "mongoose")
	selection.SetSingle(core.GroupPackageManager, "npm")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth"})
}

func TestResolveFinalRejectsInvalidCLICombination(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "mongodb")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")

	if _, err := r.ResolveFinal(selection); err == nil {
		t.Fatal("expected invalid combination error")
	}
}

func TestResolveFinalBuildsDeterministicPlan(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetMulti(core.GroupAddon, []string{"better-auth"})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	if len(plan.Fragments) == 0 {
		t.Fatal("expected fragments in the plan")
	}
	if plan.Fragments[0].Phase != core.PhaseWorkspace {
		t.Fatalf("expected first fragment phase to be workspace, got %s", plan.Fragments[0].Phase)
	}
	if len(plan.Actions) < 2 || plan.Actions[1].Command != "npx" {
		t.Fatalf("expected second action to scaffold next, got %#v", plan.Actions)
	}
}

func TestResolveFinalSupabaseWithPrismaBuildsPlan(t *testing.T) {
	reg := registry.New()
	builder := planner.New(reg)
	r := resolver.New(reg, builder)

	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "supabase")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "pnpm")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	output := builder.Format(plan)
	for _, expected := range []string{"database: supabase", "Install Prisma", "Prepare Supabase configuration"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected formatted plan to contain %q", expected)
		}
	}
}

func TestFormatPlanIncludesFragmentsAndSummary(t *testing.T) {
	reg := registry.New()
	builder := planner.New(reg)
	r := resolver.New(reg, builder)

	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "pnpm")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	output := builder.Format(plan)
	for _, expected := range []string{"Resolved template", "project: demo", "[dependencies] drizzle", "Install Drizzle"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected formatted plan to contain %q", expected)
		}
	}
}

func baseSelection() core.TemplateSelection {
	selection := core.NewTemplateSelection()
	selection.ProjectName = "demo"
	return selection
}

func testResolver() resolver.Service {
	reg := registry.New()
	return resolver.New(reg, planner.New(reg))
}

func assertChoiceIDs(t *testing.T, choices []core.ResolvedChoice, expected []string) {
	t.Helper()

	actual := make([]string, 0, len(choices))
	for _, choice := range choices {
		actual = append(actual, choice.ID)
	}
	if len(actual) != len(expected) {
		t.Fatalf("expected choices %v, got %v", expected, actual)
	}
	for _, id := range expected {
		found := false
		for _, actualID := range actual {
			if actualID == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected choice %q in %v", id, actual)
		}
	}
}
