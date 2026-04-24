package registry

import "openblueprints/internals/core"

func registerTanStackStartFrontend(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "tanstack-start",
		Name:  "TanStack Start",
		Group: core.GroupFrontend,
		Provides: []core.Capability{
			"frontend:selected",
			"frontend:tanstack-start",
			"runtime:js",
			"workspace:frontend",
			"package-manager:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{
					workspaceRootFragment(selection, "tanstack-start"),
					{
						ID:      "tanstack-start-scaffold",
						OwnerID: "tanstack-start",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							commandAction("scaffold-tanstack-start", "Scaffold TanStack Start frontend", "Creates the TanStack Start application in the apps/frontend workspace.", selection.ProjectName, "npx", "@tanstack/cli@latest", "create", "apps/frontend", "-y", "--package-manager", packageManager(selection)),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}
