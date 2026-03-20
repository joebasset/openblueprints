package registry

import "openblueprints/internals/core"

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
				if pm == "pnpm" {
					actions = append(actions,
						commandAction("express-init", "Initialize backend package", "Creates package.json for the backend workspace.", dir, "pnpm", "init"),
						commandAction("express-runtime", "Install Express runtime dependencies", "Adds the Express runtime package.", dir, "pnpm", "add", "express"),
						commandAction("express-dev", "Install TypeScript tooling", "Adds TypeScript and Node type definitions for the backend.", dir, "pnpm", "add", "-D", "typescript", "@types/express", "@types/node", "tsx"),
					)
				} else {
					actions = append(actions,
						commandAction("express-init", "Initialize backend package", "Creates package.json for the backend workspace.", dir, "npm", "init", "-y"),
						commandAction("express-runtime", "Install Express runtime dependencies", "Adds the Express runtime package.", dir, "npm", "install", "express"),
						commandAction("express-dev", "Install TypeScript tooling", "Adds TypeScript and Node type definitions for the backend.", dir, "npm", "install", "-D", "typescript", "@types/express", "@types/node", "tsx"),
					)
				}

				return []core.PlanFragment{
					{
						ID:      "express-backend",
						OwnerID: "express",
						Phase:   core.PhaseScaffold,
						Actions: actions,
					},
					{
						ID:      "express-backend-starter",
						OwnerID: "express",
						Phase:   core.PhasePostSetup,
						Actions: []core.ExecutionAction{
							noteAction("express-note", "Apply backend starter files", "MVP leaves the actual Express starter source as a follow-up patch step.", ""),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}
