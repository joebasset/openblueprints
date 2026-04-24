package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/tui"
)

type App struct {
	resolver core.Resolver
	planner  planner.Service
	tui      tui.App
}

func New(resolver core.Resolver, planBuilder planner.Service, tuiApp tui.App) App {
	return App{
		resolver: resolver,
		planner:  planBuilder,
		tui:      tuiApp,
	}
}

func (a App) Run(args []string, stdout, stderr io.Writer) error {
	selection, preview, execute, useTUI, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}

	if useTUI {
		return a.tui.Run(stdout, stderr)
	}

	if selection.ProjectName == "" {
		return errors.New("project name is required outside TUI mode; use --name or run --tui")
	}

	if err := a.completeDefaults(&selection); err != nil {
		return err
	}

	plan, err := a.resolver.ResolveFinal(selection)
	if err != nil {
		return err
	}

	if preview || !execute {
		fmt.Fprintln(stdout, a.planner.Format(plan))
	}

	if execute {
		return a.planner.Execute(context.Background(), plan, stdout, stderr)
	}

	return nil
}

func parseArgs(args []string, stderr io.Writer) (core.TemplateSelection, bool, bool, bool, error) {
	fs := flag.NewFlagSet("openblueprints", flag.ContinueOnError)
	fs.SetOutput(stderr)

	selection := core.NewTemplateSelection()
	var addons string
	var storage string
	var frontend string
	var backend string
	var database string
	var orm string
	var packageManager string
	var codeQuality string
	var finalization string
	var installSkills bool
	var preview bool
	var execute bool
	var useTUI bool

	fs.StringVar(&selection.ProjectName, "name", "", "project name")
	fs.StringVar(&frontend, "frontend", "next", "frontend pack")
	fs.StringVar(&backend, "backend", "express", "backend pack")
	fs.StringVar(&database, "database", "postgres", "database")
	fs.StringVar(&orm, "orm", "", "orm / data layer")
	fs.StringVar(&packageManager, "package-manager", "npm", "package manager")
	fs.StringVar(&codeQuality, "code-quality", "biome", "linting / formatter toolchain")
	fs.StringVar(&addons, "addons", "", "comma separated addon ids")
	fs.StringVar(&storage, "storage", "", "comma separated storage ids")
	fs.StringVar(&finalization, "finalization", "", "comma separated finalization ids")
	fs.BoolVar(&installSkills, "install-skills", false, "install selected stack agent skills")
	fs.BoolVar(&preview, "preview", false, "print the resolved plan")
	fs.BoolVar(&execute, "execute", false, "execute the resolved plan")
	fs.BoolVar(&useTUI, "tui", false, "run the interactive TUI flow")

	if err := fs.Parse(args); err != nil {
		return core.TemplateSelection{}, false, false, false, err
	}

	providedFlags := make(map[string]bool)
	fs.Visit(func(flag *flag.Flag) {
		providedFlags[flag.Name] = true
	})

	if len(args) == 0 {
		useTUI = true
	}

	selection.SetSingle(core.GroupFrontend, frontend)
	selection.SetSingle(core.GroupBackend, backend)
	selection.SetSingle(core.GroupDatabase, database)
	if orm != "" {
		selection.SetSingle(core.GroupORM, orm)
	}
	selection.SetSingle(core.GroupPackageManager, packageManager)
	selection.SetSingle(core.GroupCodeQuality, codeQuality)
	if addons != "" || providedFlags["addons"] {
		addonIDs, storageIDs := splitStorageIDs(splitCSV(addons))
		selection.SetMulti(core.GroupAddon, addonIDs)
		if len(storageIDs) > 0 {
			selection.SetMulti(core.GroupStorage, storageIDs)
		}
	}
	if storage != "" || providedFlags["storage"] {
		selection.SetMulti(core.GroupStorage, splitCSV(storage))
	}
	if finalization != "" || providedFlags["finalization"] {
		selection.SetMulti(core.GroupFinalization, removePackageInstallFinalization(splitCSV(finalization)))
	}
	if installSkills {
		selection.SetMulti(core.GroupFinalization, append(selection.Multi(core.GroupFinalization), "install-agent-skills"))
	}
	selection.Normalize()

	return selection, preview, execute, useTUI, nil
}

func removePackageInstallFinalization(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if value == "install-packages" {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
}

func splitStorageIDs(values []string) ([]string, []string) {
	addons := make([]string, 0, len(values))
	storage := make([]string, 0, len(values))
	for _, value := range values {
		switch value {
		case "s3-storage", "r2-storage":
			storage = append(storage, value)
		default:
			addons = append(addons, value)
		}
	}
	return addons, storage
}

func (a App) completeDefaults(selection *core.TemplateSelection) error {
	for {
		group, err := a.resolver.ResolveNext(*selection)
		if err != nil {
			return err
		}
		if group == nil {
			return nil
		}

		if group.Multi {
			selection.SetMulti(group.ID, defaultChoiceIDs(group.Choices))
			continue
		}

		if !group.Required {
			return fmt.Errorf("cannot complete optional %s selection non-interactively", core.ChoiceGroupLabel(group.ID))
		}

		defaultID := defaultChoiceID(group.Choices)
		if defaultID == "" {
			return fmt.Errorf("no default %s option is available", core.ChoiceGroupLabel(group.ID))
		}
		selection.SetSingle(group.ID, defaultID)
	}
}

func defaultChoiceID(choices []core.ResolvedChoice) string {
	for _, choice := range choices {
		if choice.IsDefault {
			return choice.ID
		}
	}
	if len(choices) == 0 {
		return ""
	}
	return choices[0].ID
}

func defaultChoiceIDs(choices []core.ResolvedChoice) []string {
	defaults := make([]string, 0)
	for _, choice := range choices {
		if choice.IsDefault {
			defaults = append(defaults, choice.ID)
		}
	}
	return defaults
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}
