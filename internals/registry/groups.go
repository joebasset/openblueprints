package registry

import "openblueprints/internals/core"

func registerGroups(r *Registry) {
	r.RegisterGroup(GroupDefinition{
		ID:          core.GroupFrontend,
		Title:       "Frontend pack",
		Description: "Choose the frontend foundation.",
		Required:    true,
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupBackend,
		Title:          "Backend pack",
		Description:    "Choose the backend foundation.",
		Required:       true,
		ActivationCaps: []core.Capability{"frontend:selected"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupDatabase,
		Title:          "Database",
		Description:    "Choose the backing data store.",
		Required:       true,
		ActivationCaps: []core.Capability{"backend:selected"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupStorage,
		Title:          "Storage",
		Description:    "Choose optional object storage integrations.",
		Multi:          true,
		ActivationCaps: []core.Capability{"backend:selected"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupORM,
		Title:          "ORM / data layer",
		Description:    "Choose the data access layer for the selected backend and database.",
		Required:       true,
		ActivationCaps: []core.Capability{"backend:js", "database:orm-supported"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupPackageManager,
		Title:          "Package manager",
		Description:    "Choose the JavaScript package manager used in JS workspaces.",
		Required:       true,
		ActivationCaps: []core.Capability{"runtime:js"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:             core.GroupCodeQuality,
		Title:          "Linting / formatter",
		Description:    "Choose the code quality toolchain for JavaScript workspaces.",
		Required:       true,
		ActivationCaps: []core.Capability{"runtime:js"},
	})
	r.RegisterGroup(GroupDefinition{
		ID:          core.GroupAddon,
		Title:       "Addons",
		Description: "Choose optional integrations supported by the resolved stack.",
		Multi:       true,
	})
	r.RegisterGroup(GroupDefinition{
		ID:          core.GroupFinalization,
		Title:       "Finalization",
		Description: "Choose optional finishing steps to run after scaffolding.",
		Multi:       true,
	})
}
