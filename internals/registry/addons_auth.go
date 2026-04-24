package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerAuthAddons(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "better-auth",
		Name:  "Better Auth",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:auth",
			"addon:better-auth",
		},
		RequiresAll: []core.Capability{
			"frontend:next",
			"backend:js",
			"runtime:js",
		},
		Excludes: []core.Capability{"runtime:cloudflare-workers"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install Better Auth", "Adds Better Auth to the backend workspace.", []string{"better-auth", "@better-auth/drizzle-adapter"}, nil)
				if frontendDir(selection) != backendDir(selection) {
					actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), frontendDir(selection), "Install Better Auth client", "Adds Better Auth client helpers to the Next.js frontend workspace.", []string{"better-auth"}, nil)...)
				}
				actions = append(actions,
					writeEnvAction(
						"better-auth-env",
						"Write Better Auth environment template",
						"Adds Better Auth secret and URL variables to the backend environment template.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"BETTER_AUTH_SECRET=replace-with-openssl-rand-base64-32",
							"BETTER_AUTH_URL=http://localhost:3001",
							"FRONTEND_URL=http://localhost:3000",
						},
					),
					writeFileAction("better-auth-client", "Write Better Auth frontend client", "Adds a typed frontend auth client for the Next.js app.", filepath.Join(frontendDir(selection), "src", "lib", "auth-client.ts"), betterAuthClientSource(selection)),
				)
				if !isSingleNextApp(selection) {
					actions = append(actions, writeEnvAction(
						"better-auth-frontend-env",
						"Write Better Auth frontend environment template",
						"Adds the backend auth URL for the frontend workspace.",
						frontendEnvPath(selection),
						frontendEnvLines(selection, []string{
							"BACKEND_URL=http://localhost:3001",
						}),
					))
				}
				if selection.Single(core.GroupORM) == "drizzle" {
					actions = append(actions,
						writeFileAction("better-auth-drizzle-schema", "Write Better Auth Drizzle schema", "Adds Better Auth core tables to the Drizzle schema folder.", filepath.Join(backendDir(selection), "src", "db", "schema", "auth.ts"), betterAuthDrizzleSchemaSource()),
					)
					if selection.Single(core.GroupBackend) == "next" {
						actions = append(actions,
							writeFileAction("better-auth-config", "Write Better Auth config", "Adds a Drizzle-backed Better Auth server configuration.", filepath.Join(backendDir(selection), "src", "lib", "auth.ts"), betterAuthNextDrizzleConfigSource(selection)),
							writeFileAction("better-auth-next-route", "Write Better Auth Next route", "Mounts Better Auth in the Next.js backend app router.", filepath.Join(backendDir(selection), "src", "app", "api", "auth", "[...all]", "route.ts"), betterAuthNextRouteSource()),
						)
					} else {
						actions = append(actions,
							writeFileAction("better-auth-config", "Write Better Auth config", "Adds a Drizzle-backed Better Auth server configuration.", filepath.Join(backendDir(selection), "src", "auth", "index.ts"), betterAuthNodeDrizzleConfigSource(selection)),
						)
					}
				}
				return []core.PlanFragment{{
					ID:      "better-auth-addon",
					OwnerID: "better-auth",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "better-auth/skills",
		},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "supabase-auth",
		Name:  "Supabase Auth",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:auth",
			"addon:supabase-auth",
		},
		RequiresAll: []core.Capability{
			"provider:supabase",
			"runtime:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := []core.ExecutionAction{
					writeEnvAction(
						"supabase-auth-frontend-env",
						"Write Supabase Auth frontend environment template",
						"Adds Supabase public auth variables for the frontend workspace.",
						frontendEnvPath(selection),
						frontendEnvLines(selection, []string{
							"SUPABASE_URL=https://your-project.supabase.co",
							"SUPABASE_ANON_KEY=replace-me",
						}),
					),
				}
				actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), frontendDir(selection), "Install Supabase client", "Adds the Supabase JavaScript client for auth flows.", []string{"@supabase/supabase-js"}, nil)...)
				return []core.PlanFragment{{
					ID:      "supabase-auth-addon",
					OwnerID: "supabase-auth",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "https://github.com/supabase/agent-skills",
		},
	})
}

func betterAuthClientSource(selection core.TemplateSelection) string {
	if isSingleNextApp(selection) {
		return `import { createAuthClient } from "better-auth/react";

export const authClient = createAuthClient();`
	}

	backendURLKey := "VITE_BACKEND_URL"
	if selection.Single(core.GroupFrontend) == "next" {
		backendURLKey = "NEXT_PUBLIC_BACKEND_URL"
	}
	if selection.Single(core.GroupFrontend) == "expo" {
		backendURLKey = "EXPO_PUBLIC_BACKEND_URL"
	}

	return `import { createAuthClient } from "better-auth/react";

const backendUrl = process.env.` + backendURLKey + ` || "http://localhost:3001";

export const authClient = createAuthClient({
  baseURL: backendUrl,
});`
}

func betterAuthNodeDrizzleConfigSource(selection core.TemplateSelection) string {
	appName := selection.ProjectName
	if appName == "" {
		appName = "OpenBlueprints"
	}

	return `import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";

import { db } from "../db/client";
import * as schema from "../db/schema";

const frontendUrl = process.env.FRONTEND_URL || "http://localhost:3000";

export const auth = betterAuth({
  appName: "` + appName + `",
  database: drizzleAdapter(db, {
    provider: "pg",
    schema,
  }),
  emailAndPassword: {
    enabled: true,
  },
  trustedOrigins: [frontendUrl],
});`
}

func betterAuthNextDrizzleConfigSource(selection core.TemplateSelection) string {
	appName := selection.ProjectName
	if appName == "" {
		appName = "OpenBlueprints"
	}

	return `import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";

import { db } from "../db/client";
import * as schema from "../db/schema";

const frontendUrl = process.env.FRONTEND_URL || "http://localhost:3000";

export const auth = betterAuth({
  appName: "` + appName + `",
  database: drizzleAdapter(db, {
    provider: "pg",
    schema,
  }),
  emailAndPassword: {
    enabled: true,
  },
  trustedOrigins: [frontendUrl],
});`
}

func betterAuthNextRouteSource() string {
	return `import { toNextJsHandler } from "better-auth/next-js";

import { auth } from "@/lib/auth";

export const { GET, POST } = toNextJsHandler(auth);`
}

func betterAuthDrizzleSchemaSource() string {
	return `import { boolean, pgTable, text, timestamp } from "drizzle-orm/pg-core";

export const user = pgTable("user", {
  id: text("id").primaryKey(),
  name: text("name").notNull(),
  email: text("email").notNull().unique(),
  emailVerified: boolean("email_verified").notNull(),
  image: text("image"),
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at").notNull(),
});

export const session = pgTable("session", {
  id: text("id").primaryKey(),
  token: text("token").notNull().unique(),
  userId: text("user_id").notNull().references(() => user.id, { onDelete: "cascade" }),
  expiresAt: timestamp("expires_at").notNull(),
  ipAddress: text("ip_address"),
  userAgent: text("user_agent"),
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at").notNull(),
});

export const account = pgTable("account", {
  id: text("id").primaryKey(),
  accountId: text("account_id").notNull(),
  providerId: text("provider_id").notNull(),
  userId: text("user_id").notNull().references(() => user.id, { onDelete: "cascade" }),
  accessToken: text("access_token"),
  refreshToken: text("refresh_token"),
  idToken: text("id_token"),
  accessTokenExpiresAt: timestamp("access_token_expires_at"),
  refreshTokenExpiresAt: timestamp("refresh_token_expires_at"),
  scope: text("scope"),
  password: text("password"),
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at").notNull(),
});

export const verification = pgTable("verification", {
  id: text("id").primaryKey(),
  identifier: text("identifier").notNull(),
  value: text("value").notNull(),
  expiresAt: timestamp("expires_at").notNull(),
  createdAt: timestamp("created_at"),
  updatedAt: timestamp("updated_at"),
});`
}
