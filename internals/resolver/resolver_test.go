package resolver_test

import (
	"strings"
	"testing"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
	"openblueprints/internals/resolver"
)

func TestResolveNextIncludesFrontendOptions(t *testing.T) {
	r := testResolver()
	selection := baseSelection()

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupFrontend {
		t.Fatalf("expected frontend group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"next", "expo", "tanstack-start"})
}

func TestResolveNextMongoExcludesPostgresORMs(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "mongodb")
	selection.SetMulti(core.GroupStorage, []string{})

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
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupORM {
		t.Fatalf("expected orm group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"drizzle", "prisma"})
}

func TestResolveNextBackendIncludesHono(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupBackend {
		t.Fatalf("expected backend group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"express", "hono", "hono-cf-workers", "next", "go-api"})
}

func TestResolveNextHonoPostgresIncludesExpectedORMs(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetMulti(core.GroupStorage, []string{})

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
	selection.SetMulti(core.GroupStorage, []string{})

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
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupPackageManager {
		t.Fatalf("expected package manager group, got %#v", group)
	}
}

func TestResolveNextGoShowsStorageAddonsOnly(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "go-api")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupStorage {
		t.Fatalf("expected storage group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"s3-storage", "r2-storage"})
}

func TestResolveNextAddonShownForPrismaPath(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth", "resend"})
}

func TestResolveNextAddonShownForMongoPath(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "mongodb")
	selection.SetSingle(core.GroupORM, "mongoose")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth", "resend"})
}

func TestResolveNextAddonShownForHonoPath(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth", "resend"})
}

func TestResolveNextShowsFinalizationGroupAfterAddons(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupAddon, []string{})
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupFinalization {
		t.Fatalf("expected finalization group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"git-init", "install-agent-skills"})
}

func TestResolveNextSupabaseShowsSupabaseAuthAddon(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "supabase")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"better-auth", "supabase-auth", "resend"})
}

func TestResolveNextFirebaseSkipsORM(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "firebase")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupPackageManager {
		t.Fatalf("expected package manager group after Firebase, got %#v", group)
	}
}

func TestResolveNextPackageManagersIncludeYarnAndBun(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupPackageManager {
		t.Fatalf("expected package manager group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"npm", "pnpm", "yarn", "bun"})
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
	selection.SetSingle(core.GroupCodeQuality, "biome")
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
	assertActionID(t, plan.Actions, "scaffold-next")
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
	selection.SetSingle(core.GroupCodeQuality, "biome")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	output := builder.Format(plan)
	if !strings.Contains(output, "Database         supabase") {
		t.Fatalf("expected formatted summary to contain Supabase database, got:\n%s", output)
	}
	assertActionID(t, plan.Actions, "install-prisma")
	assertActionID(t, plan.Actions, "supabase-backend-env")
	assertActionID(t, plan.Actions, "install-supabase-client")
}

func TestFormatPlanIncludesSummaryWithoutFragments(t *testing.T) {
	reg := registry.New()
	builder := planner.New(reg)
	r := resolver.New(reg, builder)

	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "pnpm")
	selection.SetSingle(core.GroupCodeQuality, "biome")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	output := builder.Format(plan)
	for _, expected := range []string{"Resolved template", "Project          demo", "ORM              drizzle"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected formatted plan to contain %q", expected)
		}
	}
	for _, unexpected := range []string{"Plan fragments", "[dependencies] drizzle", "Install Drizzle"} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("expected formatted plan to omit %q, got:\n%s", unexpected, output)
		}
	}
	assertActionPathAndKind(t, plan.Actions, "demo/pnpm-workspace.yaml", core.ActionKindWriteFile)
}

func TestResolveFinalDrizzlePostgresIncludesGeneratedFiles(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	expectedByPath := map[string]core.ActionKind{
		"demo/package.json":                        core.ActionKindWriteFile,
		"demo/packages/shared/package.json":        core.ActionKindWriteFile,
		"demo/packages/shared/src/index.ts":        core.ActionKindWriteFile,
		"demo/apps/backend/tsconfig.json":          core.ActionKindWriteFile,
		"demo/apps/backend/src/server.ts":          core.ActionKindWriteFile,
		"demo/apps/backend/.env.example":           core.ActionKindWriteEnv,
		"demo/apps/backend/drizzle.config.ts":      core.ActionKindWriteFile,
		"demo/apps/backend/src/db/client.ts":       core.ActionKindWriteFile,
		"demo/apps/backend/src/db/schema/index.ts": core.ActionKindWriteFile,
	}

	for expectedPath, expectedKind := range expectedByPath {
		assertActionPathAndKind(t, plan.Actions, expectedPath, expectedKind)
	}
	assertActionContentContains(t, plan.Actions, "workspace-package-json", `"apps/*"`)
	assertActionContentContains(t, plan.Actions, "workspace-package-json", `"packages/*"`)
}

func TestResolveFinalAlwaysInstallsPackagesWhenFinalizationOmitsInstall(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupFinalization, []string{"git-init"})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionMissingArg(t, plan.Actions, "scaffold-next", "--skip-install")
	assertActionID(t, plan.Actions, "install-express")
	assertActionMissingID(t, plan.Actions, "install-express-manual")
	assertActionID(t, plan.Actions, "install-drizzle")
	assertActionID(t, plan.Actions, "git-init")
}

func TestResolveFinalAlwaysRunsPrismaInit(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "pnpm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupFinalization, []string{})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionID(t, plan.Actions, "prisma-init")
	assertActionID(t, plan.Actions, "install-prisma")
	assertActionMissingID(t, plan.Actions, "install-prisma-manual")
}

func TestResolveFinalGoPostgresIncludesGoSpecificFiles(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "go-api")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/main.go", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/internal/database/postgres.go", core.ActionKindWriteFile)
	assertActionID(t, plan.Actions, "go-postgres-driver")
	assertActionMissingID(t, plan.Actions, "drizzle-config")
}

func TestResolveFinalHonoPostgresIncludesHonoFiles(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/tsconfig.json", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/server.ts", core.ActionKindWriteFile)
	assertActionID(t, plan.Actions, "install-hono")
	assertActionMissingID(t, plan.Actions, "install-express")
}

func TestResolveFinalHonoSupabaseDrizzleBetterAuthR2IncludesBackendPackagesAndFiles(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono")
	selection.SetSingle(core.GroupDatabase, "supabase")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "pnpm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupAddon, []string{"better-auth"})
	selection.SetMulti(core.GroupStorage, []string{"r2-storage"})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionContainsArg(t, plan.Actions, "install-hono", "hono")
	assertActionContainsArg(t, plan.Actions, "install-supabase-client", "@supabase/supabase-js")
	assertActionContainsArg(t, plan.Actions, "install-cloudflare-r2-storage-sdk", "@aws-sdk/client-s3")
	assertActionContainsArg(t, plan.Actions, "install-better-auth", "better-auth")
	assertActionContainsArg(t, plan.Actions, "install-better-auth", "@better-auth/drizzle-adapter")
	assertActionContainsArg(t, plan.Actions, "install-better-auth-client", "better-auth")
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/supabase/client.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/auth/index.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/db/schema/auth.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/frontend/src/lib/auth-client.ts", core.ActionKindWriteFile)
	assertActionContentContains(t, plan.Actions, "hono-server", "/api/auth/*")
	assertActionContentContains(t, plan.Actions, "drizzle-schema-index", `export * from "./auth";`)
	assertActionContentContains(t, plan.Actions, "supabase-backend-env", "BETTER_AUTH_SECRET=replace-with-openssl-rand-base64-32")
	assertActionContentContains(t, plan.Actions, "supabase-backend-env", "R2_BUCKET=replace-me")
	assertActionContentContains(t, plan.Actions, "better-auth-frontend-env", "NEXT_PUBLIC_BACKEND_URL=http://localhost:3001")
}

func TestResolveFinalNextBackendBetterAuthWritesNextRoute(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "next")
	selection.SetSingle(core.GroupDatabase, "supabase")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "pnpm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupAddon, []string{"better-auth"})
	selection.SetMulti(core.GroupStorage, []string{})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionID(t, plan.Actions, "scaffold-next")
	assertActionMissingID(t, plan.Actions, "scaffold-next-backend")
	assertActionContainsArg(t, plan.Actions, "install-better-auth", "better-auth")
	assertActionMissingID(t, plan.Actions, "install-better-auth-client")
	assertActionPathAndKind(t, plan.Actions, "demo/src/lib/auth.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/src/lib/auth-client.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/src/app/api/auth/[...all]/route.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/AGENTS.md", core.ActionKindWriteFile)
	assertActionMissingPathPrefix(t, plan.Actions, "demo/apps/")
	assertActionMissingPathPrefix(t, plan.Actions, "demo/packages/")
	assertActionContentContains(t, plan.Actions, "better-auth-next-route", "toNextJsHandler")
	assertActionContentContains(t, plan.Actions, "better-auth-client", "createAuthClient();")
	assertActionMissingID(t, plan.Actions, "better-auth-frontend-env")
}

func TestResolveFinalIncludesGeneratedAgentsFile(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "go-api")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupFinalization, []string{})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionPathAndKind(t, plan.Actions, "demo/AGENTS.md", core.ActionKindWriteFile)
	assertActionContentContains(t, plan.Actions, "workspace-agents", "Backend: go-api")
	assertActionContentContains(t, plan.Actions, "workspace-agents", "Do not use React useEffect")
	assertActionMissingID(t, plan.Actions, "workspace-stack-skill")
}

func TestResolveFinalInstallsSelectedAgentSkillsWhenRequested(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "supabase")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupAddon, []string{"better-auth", "supabase-auth"})
	selection.SetMulti(core.GroupStorage, []string{"r2-storage"})
	selection.SetMulti(core.GroupFinalization, []string{"install-agent-skills"})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionCommandArgs(t, plan.Actions, "install-skill-cloudflare-skills", []string{"skills", "add", "https://github.com/cloudflare/skills", "--agent", "codex", "--yes"})
	assertActionCommandArgs(t, plan.Actions, "install-skill-better-auth-skills", []string{"skills", "add", "better-auth/skills", "--agent", "codex", "--yes"})
	assertActionCommandArgs(t, plan.Actions, "install-skill-supabase-agent-skills", []string{"skills", "add", "https://github.com/supabase/agent-skills", "--agent", "codex", "--yes"})
	assertActionCommandArgs(t, plan.Actions, "install-skill-vercel-labs-agent-skills", []string{"skills", "add", "https://github.com/vercel-labs/agent-skills", "--agent", "codex", "--yes"})
}

func TestResolveFinalStorageMergesBackendEnvActions(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "express")
	selection.SetSingle(core.GroupDatabase, "postgres")
	selection.SetSingle(core.GroupORM, "prisma")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{"s3-storage", "r2-storage"})
	selection.SetMulti(core.GroupFinalization, []string{})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionContentContains(t, plan.Actions, "postgres-env", "DATABASE_URL=postgres://postgres:postgres@localhost:5432/app")
	assertActionContentContains(t, plan.Actions, "postgres-env", "S3_BUCKET=replace-me")
	assertActionContentContains(t, plan.Actions, "postgres-env", "R2_BUCKET=replace-me")
}

func TestResolveNextHonoCloudflareWorkersOnlyShowsNeonDatabase(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono-cf-workers")

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupDatabase {
		t.Fatalf("expected database group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"neon"})
}

func TestResolveNextHonoCloudflareWorkersAddons(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono-cf-workers")
	selection.SetSingle(core.GroupDatabase, "neon")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "npm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{})

	group, err := r.ResolveNext(selection)
	if err != nil {
		t.Fatalf("ResolveNext() error = %v", err)
	}
	if group == nil || group.ID != core.GroupAddon {
		t.Fatalf("expected addon group, got %#v", group)
	}
	assertChoiceIDs(t, group.Choices, []string{"resend", "cloudflare-kv", "cloudflare-queues"})
}

func TestResolveFinalHonoCloudflareWorkersIncludesWranglerBindingsAndAddons(t *testing.T) {
	r := testResolver()
	selection := baseSelection()
	selection.SetSingle(core.GroupFrontend, "next")
	selection.SetSingle(core.GroupBackend, "hono-cf-workers")
	selection.SetSingle(core.GroupDatabase, "neon")
	selection.SetSingle(core.GroupORM, "drizzle")
	selection.SetSingle(core.GroupPackageManager, "pnpm")
	selection.SetSingle(core.GroupCodeQuality, "biome")
	selection.SetMulti(core.GroupStorage, []string{"r2-storage"})
	selection.SetMulti(core.GroupAddon, []string{"cloudflare-kv", "cloudflare-queues", "resend"})
	selection.SetMulti(core.GroupFinalization, []string{"install-agent-skills"})

	plan, err := r.ResolveFinal(selection)
	if err != nil {
		t.Fatalf("ResolveFinal() error = %v", err)
	}

	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/wrangler.jsonc", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/index.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/storage/r2.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/cloudflare/kv.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/cloudflare/queue.ts", core.ActionKindWriteFile)
	assertActionPathAndKind(t, plan.Actions, "demo/apps/backend/src/email/resend.ts", core.ActionKindWriteFile)
	assertActionContainsArg(t, plan.Actions, "install-hono-workers-runtime-dev", "wrangler")
	assertActionContainsArg(t, plan.Actions, "install-drizzle", "@neondatabase/serverless")
	assertActionContainsArg(t, plan.Actions, "install-resend", "resend")
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-wrangler", `"r2_buckets"`)
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-wrangler", `"kv_namespaces"`)
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-wrangler", `"queues"`)
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-server", "R2_BUCKET: R2Bucket;")
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-server", "APP_KV: KVNamespace;")
	assertActionContentContains(t, plan.Actions, "hono-cf-workers-server", "APP_QUEUE: Queue;")
	assertActionCommandArgs(t, plan.Actions, "install-skill-cloudflare-skills", []string{"skills", "add", "https://github.com/cloudflare/skills", "--agent", "codex", "--yes"})
	assertActionCommandArgs(t, plan.Actions, "install-skill-resend-resend-skills", []string{"skills", "add", "resend/resend-skills", "--agent", "codex", "--yes"})
	assertActionCommandArgs(t, plan.Actions, "install-skill-yusukebe-hono-skill", []string{"skills", "add", "https://github.com/yusukebe/hono-skill", "--agent", "codex", "--yes"})
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

func assertActionPathAndKind(t *testing.T, actions []core.ExecutionAction, expectedPath string, expectedKind core.ActionKind) {
	t.Helper()

	for _, action := range actions {
		if action.Path != expectedPath {
			continue
		}
		if action.Kind != expectedKind {
			t.Fatalf("expected action %q to have kind %q, got %q", expectedPath, expectedKind, action.Kind)
		}
		return
	}

	t.Fatalf("expected action for path %q", expectedPath)
}

func assertActionMissingPathPrefix(t *testing.T, actions []core.ExecutionAction, unexpectedPrefix string) {
	t.Helper()

	for _, action := range actions {
		if strings.HasPrefix(action.Path, unexpectedPrefix) {
			t.Fatalf("did not expect action path %q with prefix %q", action.Path, unexpectedPrefix)
		}
	}
}

func assertActionContainsArg(t *testing.T, actions []core.ExecutionAction, actionID string, expectedArg string) {
	t.Helper()

	for _, action := range actions {
		if action.ID != actionID {
			continue
		}
		for _, arg := range action.Args {
			if arg == expectedArg {
				return
			}
		}
		t.Fatalf("expected action %q args to contain %q, got %v", actionID, expectedArg, action.Args)
	}

	t.Fatalf("expected action %q", actionID)
}

func assertActionMissingArg(t *testing.T, actions []core.ExecutionAction, actionID string, unexpectedArg string) {
	t.Helper()

	for _, action := range actions {
		if action.ID != actionID {
			continue
		}
		for _, arg := range action.Args {
			if arg == unexpectedArg {
				t.Fatalf("did not expect action %q args to contain %q, got %v", actionID, unexpectedArg, action.Args)
			}
		}
		return
	}

	t.Fatalf("expected action %q", actionID)
}

func assertActionMissingID(t *testing.T, actions []core.ExecutionAction, unexpectedID string) {
	t.Helper()

	for _, action := range actions {
		if action.ID == unexpectedID {
			t.Fatalf("did not expect action id %q in plan", unexpectedID)
		}
	}
}

func assertActionID(t *testing.T, actions []core.ExecutionAction, expectedID string) {
	t.Helper()

	for _, action := range actions {
		if action.ID == expectedID {
			return
		}
	}

	t.Fatalf("expected action id %q", expectedID)
}

func assertActionContentContains(t *testing.T, actions []core.ExecutionAction, expectedID string, expectedContent string) {
	t.Helper()

	for _, action := range actions {
		if action.ID != expectedID {
			continue
		}
		if !strings.Contains(action.Content, expectedContent) {
			t.Fatalf("expected action %q content to contain %q, got:\n%s", expectedID, expectedContent, action.Content)
		}
		return
	}

	t.Fatalf("expected action id %q", expectedID)
}

func assertActionCommandArgs(t *testing.T, actions []core.ExecutionAction, expectedID string, expectedArgs []string) {
	t.Helper()

	for _, action := range actions {
		if action.ID != expectedID {
			continue
		}
		if action.Command != "npx" {
			t.Fatalf("expected action %q command to be npx, got %q", expectedID, action.Command)
		}
		if len(action.Args) != len(expectedArgs) {
			t.Fatalf("expected action %q args %v, got %v", expectedID, expectedArgs, action.Args)
		}
		for index, expectedArg := range expectedArgs {
			if action.Args[index] != expectedArg {
				t.Fatalf("expected action %q args %v, got %v", expectedID, expectedArgs, action.Args)
			}
		}
		return
	}

	t.Fatalf("expected action id %q", expectedID)
}
