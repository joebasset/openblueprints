package registry

import "openblueprints/internals/core"

func registerDatabases(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "postgres",
		Name:      "PostgreSQL",
		Group:     core.GroupDatabase,
		IsDefault: true,
		Provides: []core.Capability{
			"database:selected",
			"database:sql",
			"database:postgres",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "postgres-config",
					OwnerID: "postgres",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						noteAction("postgres-note", "Prepare PostgreSQL configuration", "MVP leaves environment and connection file generation as a follow-up patch step.", backendDir(selection)),
					},
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "mongodb",
		Name:  "MongoDB",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:nosql",
			"database:mongodb",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "mongodb-config",
					OwnerID: "mongodb",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						noteAction("mongodb-note", "Prepare MongoDB configuration", "MVP leaves environment and connection file generation as a follow-up patch step.", backendDir(selection)),
					},
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "supabase",
		Name:  "Supabase",
		Group: core.GroupDatabase,
		Provides: []core.Capability{
			"database:selected",
			"database:sql",
			"database:postgres",
			"database:supabase",
			"provider:supabase",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "supabase-config",
					OwnerID: "supabase",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						noteAction("supabase-note", "Prepare Supabase configuration", "MVP leaves Supabase environment and client setup as a follow-up patch step.", backendDir(selection)),
					},
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})
}
