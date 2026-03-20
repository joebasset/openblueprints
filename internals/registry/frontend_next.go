package registry

import "openblueprints/internals/core"

func registerNextFrontend(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "next",
		Name:      "Next.js",
		Group:     core.GroupFrontend,
		IsDefault: true,
		Provides: []core.Capability{
			"frontend:selected",
			"frontend:next",
			"runtime:js",
			"workspace:frontend",
			"package-manager:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				args := []string{
					"create-next-app@latest",
					"frontend",
					"--ts",
					"--eslint",
					"--app",
					"--src-dir",
					"--import-alias",
					"@/*",
					"--yes",
				}
				if packageManager(selection) == "pnpm" {
					args = append(args, "--use-pnpm")
				} else {
					args = append(args, "--use-npm")
				}

				return []core.PlanFragment{
					workspaceRootFragment(selection, "next"),
					{
						ID:      "next-scaffold",
						OwnerID: "next",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							commandAction("scaffold-next", "Scaffold Next.js frontend", "Creates the Next.js application in the frontend workspace.", selection.ProjectName, "npx", args...),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}
