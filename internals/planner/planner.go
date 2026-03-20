package planner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"openblueprints/internals/core"
	"openblueprints/internals/registry"
)

type Service struct {
	registry registry.Registry
}

func New(reg registry.Registry) Service {
	return Service{registry: reg}
}

func (s Service) Build(selection core.TemplateSelection) (core.TemplatePlan, error) {
	selection.Normalize()

	entries, err := s.registry.SelectedEntries(selection)
	if err != nil {
		return core.TemplatePlan{}, err
	}

	fragments := make([]core.PlanFragment, 0)
	for _, entry := range entries {
		for _, builder := range entry.Fragments {
			if builder == nil {
				continue
			}
			fragments = append(fragments, builder(selection)...)
		}
	}

	slices.SortStableFunc(fragments, func(a, b core.PlanFragment) int {
		return slices.Index(core.OrderedPlanPhases, a.Phase) - slices.Index(core.OrderedPlanPhases, b.Phase)
	})

	actions := make([]core.ExecutionAction, 0)
	for _, fragment := range fragments {
		actions = append(actions, fragment.Actions...)
	}

	return core.TemplatePlan{
		Selection: selection,
		Fragments: fragments,
		Actions:   actions,
	}, nil
}

func (s Service) Format(plan core.TemplatePlan) string {
	var b strings.Builder
	b.WriteString("Resolved template\n")
	for _, line := range plan.Selection.SummaryLines() {
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\nPlan fragments\n")
	for i, fragment := range plan.Fragments {
		b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, fragment.Phase, fragment.OwnerID))
		for _, action := range fragment.Actions {
			b.WriteString("   - ")
			b.WriteString(action.Name)
			b.WriteString("\n")
			if action.Description != "" {
				b.WriteString("     ")
				b.WriteString(action.Description)
				b.WriteString("\n")
			}
			b.WriteString("     ")
			b.WriteString(renderAction(action))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (s Service) Execute(ctx context.Context, plan core.TemplatePlan, stdout, stderr io.Writer) error {
	baseDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	for _, action := range plan.Actions {
		fmt.Fprintf(stdout, "-> %s\n", action.Name)
		switch action.Kind {
		case core.ActionKindMkdir:
			target := filepath.Join(baseDir, action.Path)
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create directory for %s: %w", action.Name, err)
			}
		case core.ActionKindCommand:
			dir := baseDir
			if action.Dir != "" {
				dir = filepath.Join(baseDir, action.Dir)
			}
			cmd := exec.CommandContext(ctx, action.Command, action.Args...)
			cmd.Dir = dir
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("run %s: %w", action.Name, err)
			}
		case core.ActionKindNote:
			fmt.Fprintf(stdout, "   note: %s\n", action.Description)
		default:
			return fmt.Errorf("unsupported action kind %q", action.Kind)
		}
	}

	return nil
}

func renderAction(action core.ExecutionAction) string {
	switch action.Kind {
	case core.ActionKindMkdir:
		return fmt.Sprintf("mkdir -p %s", action.Path)
	case core.ActionKindNote:
		return "manual follow-up"
	case core.ActionKindCommand:
		parts := append([]string{action.Command}, action.Args...)
		command := strings.Join(parts, " ")
		if action.Dir != "" {
			return fmt.Sprintf("(cd %s && %s)", action.Dir, command)
		}
		return command
	default:
		return string(action.Kind)
	}
}
