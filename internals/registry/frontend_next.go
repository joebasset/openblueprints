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
				appTarget := "apps/frontend"
				commandDir := selection.ProjectName
				description := "Creates the Next.js application in the apps/frontend workspace."
				if isSingleNextApp(selection) {
					appTarget = selection.ProjectName
					commandDir = ""
					description = "Creates the full-stack Next.js application at the project root."
				}

				args := []string{
					"create-next-app@latest",
					appTarget,
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

				fragments := []core.PlanFragment{
					workspaceRootFragment(selection, "next"),
					{
						ID:      "next-scaffold",
						OwnerID: "next",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							commandAction("scaffold-next", "Scaffold Next.js app", description, commandDir, "npx", args...),
						},
					},
				}
				if isSingleNextApp(selection) {
					fragments = append(fragments, core.PlanFragment{
						ID:      "next-fullstack-files",
						OwnerID: "next",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("workspace-agents", "Write AGENTS.md", "Adds repo-local agent instructions for the generated stack.", selection.ProjectName+"/AGENTS.md", generatedAGENTS(selection)),
						},
					})
				}
				return fragments
			},
		},
		Properties: map[string]string{
			"kind":        "pack",
			"skillSource": "https://github.com/vercel-labs/agent-skills",
		},
	})
}
