package registry

import "openblueprints/internals/core"

func registerBackendGo(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "go-api",
		Name:  "Go API",
		Group: core.GroupBackend,
		Provides: []core.Capability{
			"backend:selected",
			"backend:go-api",
			"backend:go",
			"workspace:backend",
		},
		RequiresAll: []core.Capability{"frontend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				return []core.PlanFragment{
					{
						ID:      "go-backend",
						OwnerID: "go-api",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							{
								ID:          "go-dir",
								Name:        "Create backend workspace",
								Description: "Creates the Go backend directory.",
								Kind:        core.ActionKindMkdir,
								Path:        dir,
							},
							commandAction("go-init", "Initialize Go module", "Creates a Go module for the backend workspace.", dir, "go", "mod", "init", "backend"),
						},
					},
					{
						ID:      "go-backend-starter",
						OwnerID: "go-api",
						Phase:   core.PhasePostSetup,
						Actions: []core.ExecutionAction{
							noteAction("go-note", "Apply Go API starter files", "MVP leaves the actual Go HTTP starter source as a follow-up patch step.", ""),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}
