package registry

import "openblueprints/internals/core"

func registerBackendNext(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "next",
		Name:  "Next.js backend",
		Group: core.GroupBackend,
		Provides: []core.Capability{
			"backend:selected",
			"backend:next",
			"backend:nextjs",
			"backend:js",
			"workspace:backend",
			"package-manager:js",
		},
		RequiresAll: []core.Capability{"frontend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				if isSingleNextApp(selection) {
					return nil
				}

				args := []string{
					"create-next-app@latest",
					"apps/backend",
					"--ts",
					"--eslint",
					"--app",
					"--src-dir",
					"--import-alias",
					"@/*",
					"--disable-git",
					"--yes",
				}
				args = append(args, packageManagerFlag(selection))

				return []core.PlanFragment{{
					ID:      "next-backend-scaffold",
					OwnerID: "next",
					Phase:   core.PhaseScaffold,
					Actions: []core.ExecutionAction{
						commandAction("scaffold-next-backend", "Scaffold Next.js backend", "Creates the Next.js backend application in the apps/backend workspace.", selection.ProjectName, "npx", args...),
					},
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "pack",
			"skillSource": "https://github.com/vercel-labs/agent-skills",
		},
	})
}
