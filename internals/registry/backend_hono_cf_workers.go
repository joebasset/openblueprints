package registry

import (
	"fmt"
	"path/filepath"
	"strings"

	"openblueprints/internals/core"
)

func registerBackendHonoCloudflareWorkers(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "hono-cf-workers",
		Name:  "Hono Cloudflare Workers API",
		Group: core.GroupBackend,
		Provides: []core.Capability{
			"backend:selected",
			"backend:hono",
			"backend:hono-cf-workers",
			"backend:js",
			"runtime:cloudflare-workers",
			"provider:cloudflare",
			"workspace:backend",
		},
		RequiresAll: []core.Capability{"frontend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				pm := packageManager(selection)
				actions := []core.ExecutionAction{
					{
						ID:          "hono-cf-workers-dir",
						Name:        "Create Worker backend workspace",
						Description: "Creates the Hono Cloudflare Workers backend directory.",
						Kind:        core.ActionKindMkdir,
						Path:        dir,
					},
					writeFileAction("hono-cf-workers-package-json", "Write Worker backend package.json", "Adds scripts for Wrangler development and deployment.", filepath.Join(dir, "package.json"), honoCloudflareWorkersPackageJSON()),
				}
				actions = append(actions, packageManagerInstallActions(selection, pm, dir, "Install Hono Workers runtime", "Adds Hono and Wrangler tooling for Cloudflare Workers.", []string{"hono"}, []string{"wrangler", "typescript", "@cloudflare/workers-types"})...)

				return []core.PlanFragment{
					{
						ID:      "hono-cf-workers-backend",
						OwnerID: "hono-cf-workers",
						Phase:   core.PhaseScaffold,
						Actions: actions,
					},
					{
						ID:      "hono-cf-workers-files",
						OwnerID: "hono-cf-workers",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("hono-cf-workers-tsconfig", "Write Worker TypeScript config", "Adds TypeScript configuration for Cloudflare Workers.", filepath.Join(dir, "tsconfig.json"), honoCloudflareWorkersTSConfig()),
							writeFileAction("hono-cf-workers-wrangler", "Write Wrangler config", "Adds Cloudflare Worker, R2, KV, and Queue bindings for the selected addons.", filepath.Join(dir, "wrangler.jsonc"), honoCloudflareWorkersWranglerJSONC(selection)),
							writeFileAction("hono-cf-workers-server", "Write Worker entrypoint", "Adds a typed Hono Worker starter.", filepath.Join(dir, "src", "index.ts"), honoCloudflareWorkersSource(selection)),
						},
					},
				}
			},
		},
		Properties: map[string]string{
			"kind":         "pack",
			"skillSources": "https://github.com/cloudflare/skills,https://github.com/yusukebe/hono-skill",
		},
	})
}

func honoCloudflareWorkersPackageJSON() string {
	return fmt.Sprintf(`{
  "name": %q,
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "wrangler dev",
    "deploy": "wrangler deploy",
    "typecheck": "tsc --noEmit"
  }
}`, "backend")
}

func honoCloudflareWorkersTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "lib": ["ES2022"],
    "types": ["@cloudflare/workers-types"],
    "strict": true,
    "skipLibCheck": true,
    "noEmit": true
  },
  "include": ["src/**/*.ts", "worker-configuration.d.ts"]
}`
}

func honoCloudflareWorkersWranglerJSONC(selection core.TemplateSelection) string {
	var lines []string
	lines = append(lines,
		"{",
		`  "$schema": "node_modules/wrangler/config-schema.json",`,
		fmt.Sprintf(`  "name": %q,`, selection.ProjectName+"-api"),
		`  "main": "src/index.ts",`,
		`  "compatibility_date": "2026-04-01",`,
		`  "observability": {`,
		`    "enabled": true`,
		`  },`,
	)

	if selectionIncludes(selection, core.GroupStorage, "r2-storage") {
		lines = append(lines,
			`  "r2_buckets": [`,
			`    {`,
			`      "binding": "R2_BUCKET",`,
			fmt.Sprintf(`      "bucket_name": %q`, selection.ProjectName+"-uploads"),
			`    }`,
			`  ],`,
		)
	}
	if selectionIncludes(selection, core.GroupAddon, "cloudflare-kv") {
		lines = append(lines,
			`  "kv_namespaces": [`,
			`    {`,
			`      "binding": "APP_KV",`,
			`      "id": "replace-with-kv-namespace-id"`,
			`    }`,
			`  ],`,
		)
	}
	if selectionIncludes(selection, core.GroupAddon, "cloudflare-queues") {
		lines = append(lines,
			`  "queues": {`,
			`    "producers": [`,
			`      {`,
			`        "binding": "APP_QUEUE",`,
			fmt.Sprintf(`        "queue": %q`, selection.ProjectName+"-jobs"),
			`      }`,
			`    ],`,
			`    "consumers": [`,
			`      {`,
			fmt.Sprintf(`        "queue": %q`, selection.ProjectName+"-jobs"),
			`      }`,
			`    ]`,
			`  },`,
		)
	}

	lines[len(lines)-1] = strings.TrimSuffix(lines[len(lines)-1], ",")
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func honoCloudflareWorkersSource(selection core.TemplateSelection) string {
	bindings := []string{}
	if selectionIncludes(selection, core.GroupStorage, "r2-storage") {
		bindings = append(bindings, "  R2_BUCKET: R2Bucket;")
	}
	if selectionIncludes(selection, core.GroupAddon, "cloudflare-kv") {
		bindings = append(bindings, "  APP_KV: KVNamespace;")
	}
	if selectionIncludes(selection, core.GroupAddon, "cloudflare-queues") {
		bindings = append(bindings, "  APP_QUEUE: Queue;")
	}
	if selectionIncludes(selection, core.GroupAddon, "resend") {
		bindings = append(bindings, "  RESEND_API_KEY: string;", "  RESEND_FROM_EMAIL: string;")
	}
	if selection.Single(core.GroupDatabase) == "neon" {
		bindings = append(bindings, "  DATABASE_URL: string;")
	}
	if len(bindings) == 0 {
		bindings = append(bindings, "  ENVIRONMENT?: string;")
	}

	return `import { Hono } from "hono";

type Env = {
  Bindings: {
` + strings.Join(bindings, "\n") + `
  };
};

const app = new Hono<Env>();

app.get("/health", (context) => {
  console.log("health check requested");
  return context.json({ status: "ok" });
});

export default app;`
}
