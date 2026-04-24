package registry

import (
	"fmt"
	"path/filepath"
	"strings"

	"openblueprints/internals/core"
)

func workspaceRootFragment(selection core.TemplateSelection, ownerID string) core.PlanFragment {
	if isSingleNextApp(selection) {
		return core.PlanFragment{
			ID:      "workspace-root",
			OwnerID: ownerID,
			Phase:   core.PhaseWorkspace,
		}
	}

	actions := []core.ExecutionAction{
		{
			ID:          "workspace-root",
			Name:        "Create monorepo root",
			Description: "Creates the top-level project directory.",
			Kind:        core.ActionKindMkdir,
			Path:        selection.ProjectName,
		},
		{
			ID:          "workspace-apps",
			Name:        "Create apps workspace",
			Description: "Creates the apps workspace directory.",
			Kind:        core.ActionKindMkdir,
			Path:        appsDir(selection),
		},
		{
			ID:          "workspace-packages",
			Name:        "Create packages workspace",
			Description: "Creates the packages workspace directory.",
			Kind:        core.ActionKindMkdir,
			Path:        packagesDir(selection),
		},
		{
			ID:          "workspace-shared-package",
			Name:        "Create shared package workspace",
			Description: "Creates the shared package workspace directory.",
			Kind:        core.ActionKindMkdir,
			Path:        sharedPackageDir(selection),
		},
		writeFileAction("workspace-package-json", "Write root workspace package.json", "Adds the monorepo package manifest and workspace globs.", filepath.Join(selection.ProjectName, "package.json"), rootPackageJSON(selection)),
		writeFileAction("workspace-shared-package-json", "Write shared package manifest", "Adds a starter shared package workspace.", filepath.Join(sharedPackageDir(selection), "package.json"), sharedPackageJSON(selection)),
		writeFileAction("workspace-shared-index", "Write shared package entrypoint", "Adds a starter shared package entrypoint.", filepath.Join(sharedPackageDir(selection), "src", "index.ts"), sharedPackageSource()),
		writeFileAction("workspace-agents", "Write AGENTS.md", "Adds repo-local agent instructions for the generated stack.", filepath.Join(selection.ProjectName, "AGENTS.md"), generatedAGENTS(selection)),
	}
	actions = append(actions, workspaceConfigActions(selection)...)

	return core.PlanFragment{
		ID:      "workspace-root",
		OwnerID: ownerID,
		Phase:   core.PhaseWorkspace,
		Actions: actions,
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

func writeFileAction(id, name, description, path, content string) core.ExecutionAction {
	return core.ExecutionAction{
		ID:          id,
		Name:        name,
		Description: description,
		Kind:        core.ActionKindWriteFile,
		Path:        path,
		Content:     ensureTrailingNewline(content),
	}
}

func writeEnvAction(id, name, description, path string, lines []string) core.ExecutionAction {
	content := strings.Join(lines, "\n")
	return core.ExecutionAction{
		ID:          id,
		Name:        name,
		Description: description,
		Kind:        core.ActionKindWriteEnv,
		Path:        path,
		Content:     ensureTrailingNewline(content),
	}
}

func packageManager(selection core.TemplateSelection) string {
	pm := selection.Single(core.GroupPackageManager)
	if pm == "" {
		return "npm"
	}
	return pm
}

func packageManagerFlag(selection core.TemplateSelection) string {
	switch packageManager(selection) {
	case "pnpm":
		return "--use-pnpm"
	case "yarn":
		return "--use-yarn"
	case "bun":
		return "--use-bun"
	default:
		return "--use-npm"
	}
}

func selectionIncludes(selection core.TemplateSelection, group core.ChoiceGroupID, value string) bool {
	for _, selectedValue := range selection.Multi(group) {
		if selectedValue == value {
			return true
		}
	}
	return false
}

func isSingleNextApp(selection core.TemplateSelection) bool {
	return selection.Single(core.GroupFrontend) == "next" && selection.Single(core.GroupBackend) == "next"
}

func backendDir(selection core.TemplateSelection) string {
	if isSingleNextApp(selection) {
		return selection.ProjectName
	}
	return filepath.Join(selection.ProjectName, "apps", "backend")
}

func frontendDir(selection core.TemplateSelection) string {
	if isSingleNextApp(selection) {
		return selection.ProjectName
	}
	return filepath.Join(selection.ProjectName, "apps", "frontend")
}

func frontendEnvPath(selection core.TemplateSelection) string {
	return filepath.Join(frontendDir(selection), ".env.example")
}

func appsDir(selection core.TemplateSelection) string {
	return filepath.Join(selection.ProjectName, "apps")
}

func packagesDir(selection core.TemplateSelection) string {
	return filepath.Join(selection.ProjectName, "packages")
}

func sharedPackageDir(selection core.TemplateSelection) string {
	return filepath.Join(packagesDir(selection), "shared")
}

func packageManagerInstallActions(selection core.TemplateSelection, pm, dir, title, description string, runtimeDeps, devDeps []string) []core.ExecutionAction {
	actions := make([]core.ExecutionAction, 0, 2)
	switch pm {
	case "pnpm":
		workspaceArgs := []string{"add"}
		workspaceDevArgs := []string{"add", "-D"}
		if dir == selection.ProjectName && !isSingleNextApp(selection) {
			workspaceArgs = append(workspaceArgs, "-w")
			workspaceDevArgs = append(workspaceDevArgs, "-w")
		}
		if len(runtimeDeps) > 0 {
			actions = append(actions, commandAction(slug(title), title, description, dir, "pnpm", append(workspaceArgs, runtimeDeps...)...))
		}
		if len(devDeps) > 0 {
			actions = append(actions, commandAction(slug(title)+"-dev", title+" tooling", "Adds development tooling for the selected integration.", dir, "pnpm", append(workspaceDevArgs, devDeps...)...))
		}
	case "yarn":
		if len(runtimeDeps) > 0 {
			actions = append(actions, commandAction(slug(title), title, description, dir, "yarn", append([]string{"add"}, runtimeDeps...)...))
		}
		if len(devDeps) > 0 {
			actions = append(actions, commandAction(slug(title)+"-dev", title+" tooling", "Adds development tooling for the selected integration.", dir, "yarn", append([]string{"add", "-D"}, devDeps...)...))
		}
	case "bun":
		if len(runtimeDeps) > 0 {
			actions = append(actions, commandAction(slug(title), title, description, dir, "bun", append([]string{"add"}, runtimeDeps...)...))
		}
		if len(devDeps) > 0 {
			actions = append(actions, commandAction(slug(title)+"-dev", title+" tooling", "Adds development tooling for the selected integration.", dir, "bun", append([]string{"add", "-d"}, devDeps...)...))
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

func packageInitAction(selection core.TemplateSelection, id, name, description, dir string) core.ExecutionAction {
	switch packageManager(selection) {
	case "pnpm":
		return commandAction(id, name, description, dir, "pnpm", "init")
	case "yarn":
		return commandAction(id, name, description, dir, "yarn", "init", "-y")
	case "bun":
		return commandAction(id, name, description, dir, "bun", "init", "-y")
	default:
		return commandAction(id, name, description, dir, "npm", "init", "-y")
	}
}

func workspaceConfigActions(selection core.TemplateSelection) []core.ExecutionAction {
	if isSingleNextApp(selection) {
		return nil
	}

	if packageManager(selection) == "pnpm" {
		return []core.ExecutionAction{
			writeFileAction("workspace-pnpm", "Write pnpm workspace config", "Adds pnpm workspace globs.", filepath.Join(selection.ProjectName, "pnpm-workspace.yaml"), pnpmWorkspaceYAML()),
		}
	}

	return nil
}

func rootPackageJSON(selection core.TemplateSelection) string {
	return fmt.Sprintf(`{
  "name": %q,
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "dev:frontend": "cd apps/frontend && %s dev",
    "dev:backend": "cd apps/backend && %s dev",
    "lint": %q,
    "format": %q
  }
}`, selection.ProjectName, packageManager(selection), packageManager(selection), lintScript(selection), formatScript(selection))
}

func lintScript(selection core.TemplateSelection) string {
	switch selection.Single(core.GroupCodeQuality) {
	case "eslint-prettier":
		return "eslint ."
	case "oxlint-oxformat":
		return "oxlint ."
	default:
		return "biome check ."
	}
}

func formatScript(selection core.TemplateSelection) string {
	switch selection.Single(core.GroupCodeQuality) {
	case "eslint-prettier":
		return "prettier --write ."
	case "oxlint-oxformat":
		return "oxformat . --write"
	default:
		return "biome format --write ."
	}
}

func sharedPackageJSON(selection core.TemplateSelection) string {
	return fmt.Sprintf(`{
  "name": "@%s/shared",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "exports": {
    ".": "./src/index.ts"
  }
}`, selection.ProjectName)
}

func sharedPackageSource() string {
	return `export const workspaceName = "shared";`
}

func pnpmWorkspaceYAML() string {
	return `packages:
  - "apps/*"
  - "packages/*"`
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

func ensureTrailingNewline(value string) string {
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, "\n") {
		return value
	}
	return value + "\n"
}

func generatedAGENTS(selection core.TemplateSelection) string {
	var b strings.Builder
	b.WriteString("# AGENTS.md\n\n")
	b.WriteString("## Project Shape\n\n")
	b.WriteString("- This repository was generated by OpenBlueprints.\n")
	if isSingleNextApp(selection) {
		b.WriteString("- This is a single full-stack Next.js application at the repository root.\n")
	} else {
		b.WriteString("- Monorepo workspaces live under `apps/*` and `packages/*`.\n")
		b.WriteString(fmt.Sprintf("- Frontend: %s in `apps/frontend/`.\n", selection.Single(core.GroupFrontend)))
		b.WriteString(fmt.Sprintf("- Backend: %s in `apps/backend/`.\n", selection.Single(core.GroupBackend)))
		b.WriteString("- Shared TypeScript package: `packages/shared/`.\n")
	}
	b.WriteString(fmt.Sprintf("- Database: %s.\n", selection.Single(core.GroupDatabase)))
	if orm := selection.Single(core.GroupORM); orm != "" {
		b.WriteString(fmt.Sprintf("- Data layer: %s.\n", orm))
	}
	if addons := selection.Multi(core.GroupAddon); len(addons) > 0 {
		b.WriteString(fmt.Sprintf("- Addons: %s.\n", strings.Join(addons, ", ")))
	}
	if storage := selection.Multi(core.GroupStorage); len(storage) > 0 {
		b.WriteString(fmt.Sprintf("- Storage: %s.\n", strings.Join(storage, ", ")))
	}
	b.WriteString("\n## Development Rules\n\n")
	b.WriteString("- Use the simplest implementation that solves the problem without over-abstracting.\n")
	b.WriteString("- Keep the code scalable and easy to customize for future features.\n")
	b.WriteString("- Follow the agreed plan; do not switch implementation direction mid-change.\n")
	b.WriteString("- Do not add retries or fallback behavior unless explicitly requested.\n")
	b.WriteString("- Remove old components and unused code when refactoring.\n")
	b.WriteString("- Add logs around externally visible backend operations.\n")
	b.WriteString("- Use clear variable, module, and package names that reflect the implementation.\n")
	b.WriteString("- Keep functions clean, but avoid unnecessary tiny functions.\n")
	b.WriteString("- Make each line do one clear thing and use intermediate variables as documentation.\n")
	b.WriteString("- Do not use React useEffect; prefer derived state, event handlers, server data APIs, or framework-native alternatives.\n")
	b.WriteString("- Use the latest package versions when adding packages or initializing projects.\n")
	b.WriteString("- Let errors fail gracefully without hidden retries or fallback paths.\n")
	b.WriteString("- Build UI that looks good, feels fast, and is easy for real users to understand.\n")
	b.WriteString("- Keep frontend and backend concerns separated unless a shared package is introduced deliberately.\n")
	b.WriteString("- Update `.env.example` whenever runtime configuration changes.\n")
	b.WriteString("- Prefer small, stack-local changes over broad rewrites.\n")
	b.WriteString("\n## Stack Notes\n\n")
	switch {
	case isSingleNextApp(selection):
		b.WriteString("- Next.js app router routes and server code live under `src/app/`.\n")
		b.WriteString("- Server helpers live under `src/lib/` and `src/db/`.\n")
	case selection.Single(core.GroupBackend) == "go-api":
		b.WriteString("- Go backend code lives in `apps/backend/` and should use `go test ./...` for verification from that directory.\n")
		b.WriteString("- Database helpers live under `apps/backend/internal/database/`.\n")
	case selection.Single(core.GroupBackend) == "hono-cf-workers":
		b.WriteString("- Hono backend code lives in `apps/backend/src/` and deploys as a Cloudflare Worker with Wrangler.\n")
		b.WriteString("- Cloudflare bindings are declared in `apps/backend/wrangler.jsonc`; use Wrangler secrets for sensitive values.\n")
	case selection.Single(core.GroupBackend) == "hono":
		b.WriteString("- Hono backend code lives in `apps/backend/src/` and runs on Node.js.\n")
	default:
		b.WriteString("- Express backend code lives in `apps/backend/src/` and runs on Node.js.\n")
	}
	return b.String()
}
