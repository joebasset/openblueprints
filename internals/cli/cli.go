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
	var frontend string
	var backend string
	var database string
	var orm string
	var packageManager string
	var preview bool
	var execute bool
	var useTUI bool

	fs.StringVar(&selection.ProjectName, "name", "", "project name")
	fs.StringVar(&frontend, "frontend", "next", "frontend pack")
	fs.StringVar(&backend, "backend", "express", "backend pack")
	fs.StringVar(&database, "database", "postgres", "database")
	fs.StringVar(&orm, "orm", "", "orm / data layer")
	fs.StringVar(&packageManager, "package-manager", "npm", "package manager")
	fs.StringVar(&addons, "addons", "", "comma separated addon ids")
	fs.BoolVar(&preview, "preview", false, "print the resolved plan")
	fs.BoolVar(&execute, "execute", false, "execute the resolved plan")
	fs.BoolVar(&useTUI, "tui", false, "run the interactive TUI flow")

	if err := fs.Parse(args); err != nil {
		return core.TemplateSelection{}, false, false, false, err
	}

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
	if addons != "" {
		selection.SetMulti(core.GroupAddon, splitCSV(addons))
	}
	selection.Normalize()

	return selection, preview, execute, useTUI, nil
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
