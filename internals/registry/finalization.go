package registry

import "openblueprints/internals/core"

func registerFinalization(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "git-init",
		Name:  "Initialize git",
		Group: core.GroupFinalization,
		Provides: []core.Capability{
			"finalization:git-init",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "git-init-finalization",
					OwnerID: "git-init",
					Phase:   core.PhasePostSetup,
					Actions: []core.ExecutionAction{
						commandAction("git-init", "Initialize Git repository", "Initializes a Git repository in the scaffold root.", selection.ProjectName, "git", "init"),
					},
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "install-agent-skills",
		Name:  "Install agent skills",
		Group: core.GroupFinalization,
		Provides: []core.Capability{
			"finalization:install-agent-skills",
		},
		Properties: map[string]string{"kind": "option"},
	})
}
