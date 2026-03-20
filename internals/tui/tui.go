package tui

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
)

type App struct {
	resolver core.Resolver
	planner  planner.Service
}

func New(resolver core.Resolver, planBuilder planner.Service) App {
	return App{
		resolver: resolver,
		planner:  planBuilder,
	}
}

func (a App) Run(stdout, stderr io.Writer) error {
	selection := core.NewTemplateSelection()

	if err := runTextInput("Project name", "Used as the workspace directory name.", &selection.ProjectName); err != nil {
		return err
	}

	for {
		group, err := a.resolver.ResolveNext(selection)
		if err != nil {
			return err
		}
		if group == nil {
			break
		}

		if group.Multi {
			var values []string
			if err := runMultiSelect(*group, &values); err != nil {
				return err
			}
			selection.SetMulti(group.ID, values)
			continue
		}

		var value string
		if err := runSelect(*group, &value); err != nil {
			return err
		}
		selection.SetSingle(group.ID, value)
	}

	plan, err := a.resolver.ResolveFinal(selection)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, a.planner.Format(plan))

	var shouldExecute bool
	if err := runConfirm("Execute scaffold plan", "Run the generated command sequence now.", &shouldExecute); err != nil {
		return err
	}

	if !shouldExecute {
		return nil
	}

	return a.planner.Execute(context.Background(), plan, stdout, stderr)
}

func runTextInput(title, description string, value *string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Description(description).
				Value(value),
		),
	)
	form.WithInput(os.Stdin)
	form.WithOutput(os.Stderr)
	return form.Run()
}

func runConfirm(title, description string, value *bool) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Value(value),
		),
	)
	form.WithInput(os.Stdin)
	form.WithOutput(os.Stderr)
	return form.Run()
}

func runSelect(group core.ChoiceGroup, value *string) error {
	options := make([]huh.Option[string], 0, len(group.Choices))
	defaultValue := ""

	for _, choice := range group.Choices {
		options = append(options, huh.NewOption(choice.Name, choice.ID))
		if choice.IsDefault && defaultValue == "" {
			defaultValue = choice.ID
		}
	}
	if defaultValue != "" {
		*value = defaultValue
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(group.Title).
				Description(group.Description).
				Options(options...).
				Value(value),
		),
	)
	form.WithInput(os.Stdin)
	form.WithOutput(os.Stderr)
	return form.Run()
}

func runMultiSelect(group core.ChoiceGroup, value *[]string) error {
	options := make([]huh.Option[string], 0, len(group.Choices))
	for _, choice := range group.Choices {
		options = append(options, huh.NewOption(choice.Name, choice.ID))
		if choice.IsDefault {
			*value = append(*value, choice.ID)
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(group.Title).
				Description(group.Description).
				Options(options...).
				Value(value),
		),
	)
	form.WithInput(os.Stdin)
	form.WithOutput(os.Stderr)
	return form.Run()
}
