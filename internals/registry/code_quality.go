package registry

import "openblueprints/internals/core"

func registerCodeQuality(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:        "biome",
		Name:      "Biome",
		Group:     core.GroupCodeQuality,
		IsDefault: true,
		Provides: []core.Capability{
			"code-quality:biome",
		},
		RequiresAll: []core.Capability{"package-manager:js"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := packageManagerInstallActions(selection, packageManager(selection), selection.ProjectName, "Install Biome", "Adds Biome linting and formatting to the workspace.", nil, []string{"@biomejs/biome"})
				actions = append(actions, writeFileAction("biome-config", "Write Biome config", "Adds a workspace Biome configuration.", selection.ProjectName+"/biome.json", biomeConfigSource()))
				return []core.PlanFragment{{
					ID:      "biome-code-quality",
					OwnerID: "biome",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "eslint-prettier",
		Name:  "ESLint + Prettier",
		Group: core.GroupCodeQuality,
		Provides: []core.Capability{
			"code-quality:eslint",
			"code-quality:prettier",
		},
		RequiresAll: []core.Capability{"package-manager:js"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := packageManagerInstallActions(selection, packageManager(selection), selection.ProjectName, "Install ESLint and Prettier", "Adds ESLint and Prettier to the workspace.", nil, []string{"eslint", "prettier", "typescript-eslint"})
				actions = append(actions,
					writeFileAction("eslint-config", "Write ESLint config", "Adds a workspace ESLint flat config.", selection.ProjectName+"/eslint.config.mjs", eslintConfigSource()),
					writeFileAction("prettier-config", "Write Prettier config", "Adds a workspace Prettier configuration.", selection.ProjectName+"/.prettierrc.json", prettierConfigSource()),
				)
				return []core.PlanFragment{{
					ID:      "eslint-prettier-code-quality",
					OwnerID: "eslint-prettier",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "oxlint-oxformat",
		Name:  "Oxlint + Oxformat",
		Group: core.GroupCodeQuality,
		Provides: []core.Capability{
			"code-quality:oxlint",
			"code-quality:oxformat",
		},
		RequiresAll: []core.Capability{"package-manager:js"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "oxlint-oxformat-code-quality",
					OwnerID: "oxlint-oxformat",
					Phase:   core.PhaseIntegration,
					Actions: packageManagerInstallActions(selection, packageManager(selection), selection.ProjectName, "Install Oxlint and Oxformat", "Adds Oxlint and Oxformat to the workspace.", nil, []string{"oxlint", "oxformat"}),
				}}
			},
		},
		Properties: map[string]string{"kind": "option"},
	})
}

func biomeConfigSource() string {
	return `{
  "$schema": "https://biomejs.dev/schemas/2.0.0/schema.json",
  "formatter": {
    "enabled": true
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true
    }
  }
}`
}

func eslintConfigSource() string {
	return `import tseslint from "typescript-eslint";

export default tseslint.config(
  ...tseslint.configs.recommended,
);`
}

func prettierConfigSource() string {
	return `{
  "semi": true,
  "singleQuote": false,
  "trailingComma": "all"
}`
}
