package registry

import "openblueprints/internals/core"

func registerExpoFrontend(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "expo",
		Name:  "React Native Expo",
		Group: core.GroupFrontend,
		Provides: []core.Capability{
			"frontend:selected",
			"frontend:expo",
			"runtime:js",
			"runtime:react-native",
			"workspace:frontend",
			"package-manager:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				args := []string{
					"create-expo-app@latest",
					"apps/frontend",
					"--yes",
					"--no-agents-md",
				}
				return []core.PlanFragment{
					workspaceRootFragment(selection, "expo"),
					{
						ID:      "expo-scaffold",
						OwnerID: "expo",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							commandAction("scaffold-expo", "Scaffold Expo frontend", "Creates the Expo application in the apps/frontend workspace.", selection.ProjectName, "npx", args...),
						},
					},
				}
			},
		},
		Properties: map[string]string{
			"kind":        "pack",
			"skillSource": "https://github.com/expo/skills",
		},
	})
}
