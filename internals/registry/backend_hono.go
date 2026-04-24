package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerBackendHono(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "hono",
		Name:  "Hono API",
		Group: core.GroupBackend,
		Provides: []core.Capability{
			"backend:selected",
			"backend:hono",
			"backend:js",
			"workspace:backend",
		},
		RequiresAll: []core.Capability{"frontend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				pm := packageManager(selection)
				dir := backendDir(selection)
				actions := []core.ExecutionAction{
					{
						ID:          "hono-dir",
						Name:        "Create backend workspace",
						Description: "Creates the Hono backend directory.",
						Kind:        core.ActionKindMkdir,
						Path:        dir,
					},
				}
				actions = append(actions, packageInitAction(selection, "hono-init", "Initialize backend package", "Creates package.json for the backend workspace.", dir))
				actions = append(actions, packageManagerInstallActions(selection, pm, dir, "Install Hono", "Adds the Hono runtime and backend TypeScript tooling.", []string{"hono", "@hono/node-server"}, []string{"typescript", "@types/node", "tsx"})...)

				return []core.PlanFragment{
					{
						ID:      "hono-backend",
						OwnerID: "hono",
						Phase:   core.PhaseScaffold,
						Actions: actions,
					},
					{
						ID:      "hono-backend-files",
						OwnerID: "hono",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("hono-tsconfig", "Write Hono TypeScript config", "Adds the backend TypeScript compiler configuration.", filepath.Join(backendDir(selection), "tsconfig.json"), nodeBackendTSConfig()),
							writeFileAction("hono-server", "Write Hono server entrypoint", "Adds a minimal typed Hono server starter.", filepath.Join(backendDir(selection), "src", "server.ts"), honoServerSource(selection)),
						},
					},
				}
			},
		},
		Properties: map[string]string{
			"kind":         "pack",
			"skillSources": "https://github.com/yusukebe/hono-skill",
		},
	})
}

func honoServerSource(selection core.TemplateSelection) string {
	if selectionIncludes(selection, core.GroupAddon, "better-auth") {
		return `import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { cors } from "hono/cors";

import { auth } from "./auth";

const app = new Hono();
const port = Number(process.env.PORT || "3001");
const frontendUrl = process.env.FRONTEND_URL || "http://localhost:3000";

app.use(
  "/api/auth/*",
  cors({
    origin: frontendUrl,
    allowHeaders: ["Content-Type", "Authorization"],
    allowMethods: ["POST", "GET", "OPTIONS"],
    credentials: true,
  }),
);

app.on(["POST", "GET"], "/api/auth/*", (context) => {
  return auth.handler(context.req.raw);
});

app.get("/health", (context) => {
  return context.json({ status: "ok" });
});

serve({ fetch: app.fetch, port }, (info) => {
  console.log("hono server listening on port", info.port);
});`
	}

	return `import { serve } from "@hono/node-server";
import { Hono } from "hono";

const app = new Hono();
const port = Number(process.env.PORT || "3001");

app.get("/health", (context) => {
  return context.json({ status: "ok" });
});

serve({ fetch: app.fetch, port }, (info) => {
  console.log("hono server listening on port", info.port);
});`
}
