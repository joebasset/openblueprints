package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func workspaceRootFragment(selection core.TemplateSelection, ownerID string) core.PlanFragment {
	return core.PlanFragment{
		ID:      "workspace-root",
		OwnerID: ownerID,
		Phase:   core.PhaseWorkspace,
		Actions: []core.ExecutionAction{
			{
				ID:          "workspace-root",
				Name:        "Create workspace root",
				Description: "Creates the top-level project directory.",
				Kind:        core.ActionKindMkdir,
				Path:        selection.ProjectName,
			},
		},
	}
}

func commandAction(id, name, description, dir, command string, args ...string) core.ExecutionAction {
	return core.ExecutionAction{
		ID:          id,
		Name:        name,
		Description: description,
		Kind:        core.ActionKindCommand,
		Dir:         dir,
		Command:     command,
		Args:        args,
	}
}

func noteAction(id, name, description, dir string) core.ExecutionAction {
	return core.ExecutionAction{
		ID:          id,
		Name:        name,
		Description: description,
		Kind:        core.ActionKindNote,
		Dir:         dir,
	}
}

func packageManager(selection core.TemplateSelection) string {
	pm := selection.Single(core.GroupPackageManager)
	if pm == "" {
		return "npm"
	}
	return pm
}

func backendDir(selection core.TemplateSelection) string {
	return filepath.Join(selection.ProjectName, "backend")
}

func frontendDir(selection core.TemplateSelection) string {
	return filepath.Join(selection.ProjectName, "frontend")
}

func packageManagerInstallActions(pm, dir, title, description string, runtimeDeps, devDeps []string) []core.ExecutionAction {
	actions := make([]core.ExecutionAction, 0, 2)
	switch pm {
	case "pnpm":
		if len(runtimeDeps) > 0 {
			actions = append(actions, commandAction(slug(title), title, description, dir, "pnpm", append([]string{"add"}, runtimeDeps...)...))
		}
		if len(devDeps) > 0 {
			actions = append(actions, commandAction(slug(title)+"-dev", title+" tooling", "Adds development tooling for the selected integration.", dir, "pnpm", append([]string{"add", "-D"}, devDeps...)...))
		}
	default:
		if len(runtimeDeps) > 0 {
			actions = append(actions, commandAction(slug(title), title, description, dir, "npm", append([]string{"install"}, runtimeDeps...)...))
		}
		if len(devDeps) > 0 {
			actions = append(actions, commandAction(slug(title)+"-dev", title+" tooling", "Adds development tooling for the selected integration.", dir, "npm", append([]string{"install", "-D"}, devDeps...)...))
		}
	}
	return actions
}

func slug(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			out = append(out, r)
		case r == ' ' || r == '-':
			out = append(out, '-')
		}
	}
	return string(out)
}
