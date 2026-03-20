package registry

import "openblueprints/internals/core"

func registerORMs(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "prisma",
		Name:      "Prisma",
		Group:     core.GroupORM,
		IsDefault: true,
		Provides: []core.Capability{
			"orm:prisma",
			"orm-family:schema",
		},
		RequiresAll: []core.Capability{
			"backend:express",
			"database:sql",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				pm := packageManager(selection)
				actions := packageManagerInstallActions(pm, dir, "Install Prisma", "Adds Prisma packages for the Express backend.", []string{"prisma", "@prisma/client"}, nil)
				command := "npx"
				args := []string{"prisma", "init"}
				if pm == "pnpm" {
					command = "pnpm"
					args = []string{"exec", "prisma", "init"}
				}
				actions = append(actions, commandAction("prisma-init", "Initialize Prisma", "Creates the initial Prisma configuration.", dir, command, args...))
				return []core.PlanFragment{{
					ID:      "prisma-orm",
					OwnerID: "prisma",
					Phase:   core.PhaseDependencies,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "drizzle",
		Name:  "Drizzle",
		Group: core.GroupORM,
		Provides: []core.Capability{
			"orm:drizzle",
			"orm-family:sql-builder",
		},
		RequiresAll: []core.Capability{
			"backend:express",
			"database:sql",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "drizzle-orm",
					OwnerID: "drizzle",
					Phase:   core.PhaseDependencies,
					Actions: packageManagerInstallActions(packageManager(selection), backendDir(selection), "Install Drizzle", "Adds Drizzle ORM packages for SQL databases.", []string{"drizzle-orm", "pg"}, []string{"drizzle-kit"}),
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "mongoose",
		Name:  "Mongoose",
		Group: core.GroupORM,
		Provides: []core.Capability{
			"orm:mongoose",
			"orm-family:document",
		},
		RequiresAll: []core.Capability{
			"backend:express",
			"database:nosql",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "mongoose-orm",
					OwnerID: "mongoose",
					Phase:   core.PhaseDependencies,
					Actions: packageManagerInstallActions(packageManager(selection), backendDir(selection), "Install Mongoose", "Adds MongoDB support for the Express backend.", []string{"mongoose"}, nil),
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})
}
