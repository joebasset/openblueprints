package registry

import "openblueprints/internals/core"

func registerPackageManagers(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "npm",
		Name:      "npm",
		Group:     core.GroupPackageManager,
		IsDefault: true,
		Provides:  []core.Capability{"pm:npm"},
		RequiresAll: []core.Capability{
			"package-manager:js",
		},
		Properties: map[string]string{"kind": "option"},
	})
	r.RegisterEntry(EntryDefinition{
		ID:    "pnpm",
		Name:  "pnpm",
		Group: core.GroupPackageManager,
		Provides: []core.Capability{
			"pm:pnpm",
		},
		RequiresAll: []core.Capability{
			"package-manager:js",
		},
		Properties: map[string]string{"kind": "option"},
	})
	r.RegisterEntry(EntryDefinition{
		ID:    "yarn",
		Name:  "Yarn",
		Group: core.GroupPackageManager,
		Provides: []core.Capability{
			"pm:yarn",
		},
		RequiresAll: []core.Capability{
			"package-manager:js",
		},
		Properties: map[string]string{"kind": "option"},
	})
	r.RegisterEntry(EntryDefinition{
		ID:    "bun",
		Name:  "Bun",
		Group: core.GroupPackageManager,
		Provides: []core.Capability{
			"pm:bun",
		},
		RequiresAll: []core.Capability{
			"package-manager:js",
		},
		Properties: map[string]string{"kind": "option"},
	})
}
