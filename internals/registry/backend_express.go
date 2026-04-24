package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerBackendExpress(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "express",
		Name:      "Express API",
		Group:     core.GroupBackend,
		IsDefault: true,
		Provides: []core.Capability{
			"backend:selected",
			"backend:express",
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
						ID:          "express-dir",
						Name:        "Create backend workspace",
						Description: "Creates the Express backend directory.",
						Kind:        core.ActionKindMkdir,
						Path:        dir,
					},
				}
				actions = append(actions, packageInitAction(selection, "express-init", "Initialize backend package", "Creates package.json for the backend workspace.", dir))
				actions = append(actions, packageManagerInstallActions(selection, pm, dir, "Install Express", "Adds the Express runtime and backend TypeScript tooling.", []string{"express"}, []string{"typescript", "@types/express", "@types/node", "tsx"})...)

				return []core.PlanFragment{
					{
						ID:      "express-backend",
						OwnerID: "express",
						Phase:   core.PhaseScaffold,
						Actions: actions,
					},
					{
						ID:      "express-backend-files",
						OwnerID: "express",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("express-tsconfig", "Write Express TypeScript config", "Adds the backend TypeScript compiler configuration.", filepath.Join(backendDir(selection), "tsconfig.json"), nodeBackendTSConfig()),
							writeFileAction("express-server", "Write Express server entrypoint", "Adds a minimal typed Express server starter.", filepath.Join(backendDir(selection), "src", "server.ts"), expressServerSource()),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}

func nodeBackendTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "rootDir": "./src",
    "outDir": "./dist",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src/**/*.ts"]
}`
}

func expressServerSource() string {
	return `import express from "express";

const app = express();
const port = process.env.PORT || "3001";

app.get("/health", (_request, response) => {
  response.json({ status: "ok" });
});

app.listen(port, () => {
  console.log("express server listening on port", port);
});`
}
