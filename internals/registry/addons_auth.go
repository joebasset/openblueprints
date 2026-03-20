package registry

import "openblueprints/internals/core"

func registerAuthAddons(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "better-auth",
		Name:  "Better Auth",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:better-auth",
		},
		RequiresAll: []core.Capability{
			"frontend:next",
			"backend:express",
			"runtime:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "better-auth-addon",
					OwnerID: "better-auth",
					Phase:   core.PhaseIntegration,
					Actions: packageManagerInstallActions(packageManager(selection), frontendDir(selection), "Install Better Auth", "Adds Better Auth to the Next.js frontend workspace.", []string{"better-auth"}, nil),
				}}
			},
		},
		Properties: map[string]string{"kind": "addon"},
	})
}
