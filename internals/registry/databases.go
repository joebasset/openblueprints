package registry

import (
	"path/filepath"
	"strings"

	"openblueprints/internals/core"
)

func registerDatabases(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "postgres",
		Name:      "PostgreSQL",
		Group:     core.GroupDatabase,
		IsDefault: true,
		Provides: []core.Capability{
			"database:selected",
			"database:orm-supported",
			"database:sql",
			"database:postgres",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Excludes:    []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := []core.ExecutionAction{
					writeEnvAction(
						"postgres-env",
						"Write PostgreSQL environment template",
						"Adds a starter PostgreSQL connection string for the backend workspace.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"DATABASE_URL=postgres://postgres:postgres@localhost:5432/app",
							"PORT=3001",
						},
					),
				}
				if selection.Single(core.GroupBackend) == "go-api" {
					actions = append(actions, writeFileAction("postgres-go-client", "Write Go PostgreSQL client", "Adds a shared PostgreSQL connection helper for the Go backend.", filepath.Join(backendDir(selection), "internal", "database", "postgres.go"), goPostgresSource()))
					actions = append(actions, commandAction("go-postgres-driver", "Install pgx PostgreSQL driver", "Adds the pgx driver for the Go backend.", backendDir(selection), "go", "get", "github.com/jackc/pgx/v5@latest"))
				}
				return []core.PlanFragment{{
					ID:      "postgres-config",
					OwnerID: "postgres",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "mongodb",
		Name:  "MongoDB",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:orm-supported",
			"database:nosql",
			"database:mongodb",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Excludes:    []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "mongodb-config",
					OwnerID: "mongodb",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						noteAction("mongodb-note", "Prepare MongoDB configuration", "MVP leaves environment and connection file generation as a follow-up patch step.", backendDir(selection)),
					},
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "supabase",
		Name:  "Supabase",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:orm-supported",
			"database:sql",
			"database:postgres",
			"database:supabase",
			"provider:supabase",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Excludes:    []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := []core.ExecutionAction{
					writeEnvAction(
						"supabase-backend-env",
						"Write Supabase backend environment template",
						"Adds Supabase service credentials and the PostgreSQL connection string for the backend workspace.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres",
							"SUPABASE_URL=https://your-project.supabase.co",
							"SUPABASE_SERVICE_ROLE_KEY=replace-me",
							"PORT=3001",
						},
					),
				}
				if selection.Single(core.GroupBackend) == "go-api" {
					actions = append(actions, writeFileAction("supabase-go-postgres-client", "Write Go PostgreSQL client", "Adds a shared PostgreSQL connection helper for the Go backend.", filepath.Join(backendDir(selection), "internal", "database", "postgres.go"), goPostgresSource()))
					actions = append(actions, commandAction("go-supabase-postgres-driver", "Install pgx PostgreSQL driver", "Adds the pgx driver for the Go backend.", backendDir(selection), "go", "get", "github.com/jackc/pgx/v5@latest"))
				} else {
					actions = append(actions,
						writeFileAction("supabase-js-client", "Write Supabase backend client", "Adds a server-side Supabase client helper for the backend workspace.", filepath.Join(backendDir(selection), "src", "supabase", "client.ts"), supabaseJSClientSource()),
					)
					actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install Supabase client", "Adds the Supabase JavaScript client to the backend workspace.", []string{"@supabase/supabase-js"}, nil)...)
				}
				return []core.PlanFragment{{
					ID:      "supabase-config",
					OwnerID: "supabase",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "option",
			"skillSource": "https://github.com/supabase/agent-skills",
		},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "neon",
		Name:  "Neon Postgres",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:orm-supported",
			"database:sql",
			"database:postgres",
			"database:neon",
			"provider:neon",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := []core.ExecutionAction{
					writeEnvAction(
						"neon-env",
						"Write Neon environment template",
						"Adds a Neon Postgres connection string for the backend workspace.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"DATABASE_URL=postgresql://user:password@ep-example.region.aws.neon.tech/app?sslmode=require",
							"PORT=3001",
						},
					),
				}
				if selection.Single(core.GroupBackend) == "go-api" {
					actions = append(actions, writeFileAction("neon-go-postgres-client", "Write Go PostgreSQL client", "Adds a shared PostgreSQL connection helper for the Go backend.", filepath.Join(backendDir(selection), "internal", "database", "postgres.go"), goPostgresSource()))
					actions = append(actions, commandAction("go-neon-postgres-driver", "Install pgx PostgreSQL driver", "Adds the pgx driver for the Go backend.", backendDir(selection), "go", "get", "github.com/jackc/pgx/v5@latest"))
				}
				return []core.PlanFragment{{
					ID:      "neon-config",
					OwnerID: "neon",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "firebase",
		Name:  "Firebase",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:firebase",
			"provider:firebase",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Excludes:    []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := []core.ExecutionAction{
					writeEnvAction(
						"firebase-backend-env",
						"Write Firebase backend environment template",
						"Adds Firebase project credentials for the backend workspace.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"FIREBASE_PROJECT_ID=replace-me",
							"FIREBASE_CLIENT_EMAIL=replace-me",
							"FIREBASE_PRIVATE_KEY=replace-me",
							"PORT=3001",
						},
					),
					writeEnvAction(
						"firebase-frontend-env",
						"Write Firebase frontend environment template",
						"Adds Firebase public client configuration for the frontend workspace.",
						frontendEnvPath(selection),
						frontendEnvLines(selection, []string{
							"FIREBASE_API_KEY=replace-me",
							"FIREBASE_AUTH_DOMAIN=replace-me",
							"FIREBASE_PROJECT_ID=replace-me",
							"FIREBASE_STORAGE_BUCKET=replace-me",
							"FIREBASE_APP_ID=replace-me",
						}),
					),
				}
				switch selection.Single(core.GroupBackend) {
				case "go-api":
					actions = append(actions, commandAction("go-firebase-admin", "Install Firebase Admin SDK", "Adds the Firebase Admin SDK for the Go backend.", backendDir(selection), "go", "get", "firebase.google.com/go/v4@latest"))
				default:
					actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install Firebase Admin", "Adds Firebase Admin SDK support for the selected JS backend.", []string{"firebase-admin"}, nil)...)
				}
				return []core.PlanFragment{{
					ID:      "firebase-config",
					OwnerID: "firebase",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})
}

func goPostgresSource() string {
	return `package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	connectionString := os.Getenv("DATABASE_URL")
	return pgxpool.New(ctx, connectionString)
}`
}

func supabaseJSClientSource() string {
	return `import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.SUPABASE_URL;
const serviceRoleKey = process.env.SUPABASE_SERVICE_ROLE_KEY;

if (!supabaseUrl) {
  throw new Error("SUPABASE_URL is required");
}

if (!serviceRoleKey) {
  throw new Error("SUPABASE_SERVICE_ROLE_KEY is required");
}

export const supabase = createClient(supabaseUrl, serviceRoleKey, {
  auth: {
    persistSession: false,
  },
});`
}

func frontendEnvLines(selection core.TemplateSelection, keys []string) []string {
	prefix := frontendEnvPrefix(selection)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		parts := strings.SplitN(key, "=", 2)
		if len(parts) != 2 {
			lines = append(lines, prefix+key)
			continue
		}
		lines = append(lines, prefix+parts[0]+"="+parts[1])
	}
	return lines
}

func frontendEnvPrefix(selection core.TemplateSelection) string {
	switch selection.Single(core.GroupFrontend) {
	case "next":
		return "NEXT_PUBLIC_"
	case "expo":
		return "EXPO_PUBLIC_"
	default:
		return "VITE_"
	}
}
