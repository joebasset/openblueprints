package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerORMs(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "prisma",
		Name:      "Prisma",
		Group:     core.GroupORM,
		IsDefault: true,
		Provides: []core.Capability{
			"orm:prisma",
			"orm-family:schema",
		},
		RequiresAll: []core.Capability{
			"backend:js",
			"database:sql",
		},
		Excludes: []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				pm := packageManager(selection)
				actions := packageManagerInstallActions(selection, pm, dir, "Install Prisma", "Adds Prisma packages for the selected JS backend.", []string{"prisma", "@prisma/client"}, nil)
				command := "npx"
				args := []string{"prisma", "init"}
				if pm == "pnpm" {
					command = "pnpm"
					args = []string{"exec", "prisma", "init"}
				}
				actions = append(actions, commandAction("prisma-init", "Initialize Prisma", "Creates the initial Prisma configuration.", dir, command, args...))
				return []core.PlanFragment{{
					ID:      "prisma-orm",
					OwnerID: "prisma",
					Phase:   core.PhaseDependencies,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "drizzle",
		Name:  "Drizzle",
		Group: core.GroupORM,
		Provides: []core.Capability{
			"orm:drizzle",
			"orm-family:sql-builder",
		},
		RequiresAll: []core.Capability{
			"backend:js",
			"database:sql",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				return []core.PlanFragment{
					{
						ID:      "drizzle-orm",
						OwnerID: "drizzle",
						Phase:   core.PhaseDependencies,
						Actions: drizzleInstallActions(selection, dir),
					},
					{
						ID:      "drizzle-files",
						OwnerID: "drizzle",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("drizzle-config", "Write Drizzle config", "Adds the Drizzle configuration file for the backend workspace.", filepath.Join(backendDir(selection), "drizzle.config.ts"), drizzleConfigSource()),
							writeFileAction("drizzle-client", "Write Drizzle database client", "Adds a shared PostgreSQL client for Drizzle-powered integrations.", filepath.Join(backendDir(selection), "src", "db", "client.ts"), drizzleClientSource(selection)),
							writeFileAction("drizzle-schema-index", "Write Drizzle schema index", "Adds the schema entrypoint used by Drizzle integrations.", filepath.Join(backendDir(selection), "src", "db", "schema", "index.ts"), drizzleSchemaIndexSource(selection)),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "mongoose",
		Name:  "Mongoose",
		Group: core.GroupORM,
		Provides: []core.Capability{
			"orm:mongoose",
			"orm-family:document",
		},
		RequiresAll: []core.Capability{
			"backend:js",
			"database:nosql",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "mongoose-orm",
					OwnerID: "mongoose",
					Phase:   core.PhaseDependencies,
					Actions: packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install Mongoose", "Adds MongoDB support for the selected JS backend.", []string{"mongoose"}, nil),
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})
}

func drizzleConfigSource() string {
	return `import { defineConfig } from "drizzle-kit";

export default defineConfig({
  schema: "./src/db/schema/index.ts",
  out: "./drizzle",
  dialect: "postgresql",
  dbCredentials: {
    url: process.env.DATABASE_URL || "",
  },
});`
}

func drizzleInstallActions(selection core.TemplateSelection, dir string) []core.ExecutionAction {
	runtimeDeps := []string{"drizzle-orm", "pg"}
	if selection.Single(core.GroupBackend) == "hono-cf-workers" && selection.Single(core.GroupDatabase) == "neon" {
		runtimeDeps = []string{"drizzle-orm", "@neondatabase/serverless"}
	}
	return packageManagerInstallActions(selection, packageManager(selection), dir, "Install Drizzle", "Adds Drizzle ORM packages for SQL databases.", runtimeDeps, []string{"drizzle-kit"})
}

func drizzleClientSource(selection core.TemplateSelection) string {
	if selection.Single(core.GroupBackend) == "hono-cf-workers" && selection.Single(core.GroupDatabase) == "neon" {
		return `import { neon } from "@neondatabase/serverless";
import { drizzle } from "drizzle-orm/neon-http";

export function createDb(connectionString: string) {
  if (!connectionString) {
    throw new Error("DATABASE_URL is required");
  }

  const sql = neon(connectionString);
  return drizzle(sql);
}`
	}

	return `import { drizzle } from "drizzle-orm/node-postgres";
import { Pool } from "pg";

const connectionString = process.env.DATABASE_URL;

if (!connectionString) {
  throw new Error("DATABASE_URL is required");
}

const pool = new Pool({ connectionString });

export const db = drizzle(pool);`
}

func drizzleSchemaIndexSource(selection core.TemplateSelection) string {
	if selectionIncludes(selection, core.GroupAddon, "better-auth") {
		return `export * from "./auth";`
	}

	return `export {};`
}
