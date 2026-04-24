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
	if shouldInstallAgentSkills(selection) {
		fragment := agentSkillsFragment(selection, entries)
		if len(fragment.Actions) > 0 {
			fragments = append(fragments, fragment)
			actions = append(actions, fragment.Actions...)
		}
	}

	return core.TemplatePlan{
		Selection: selection,
		Fragments: fragments,
		Actions:   mergeEnvWriteActions(actions),
	}, nil
}

func (s Service) Format(plan core.TemplatePlan) string {
	var b strings.Builder
	b.WriteString("Resolved template\n")
	b.WriteString("-----------------\n")
	for _, line := range plan.Selection.SummaryLines() {
		label, value, found := strings.Cut(line, ": ")
		if !found {
			b.WriteString(line)
			b.WriteString("\n")
			continue
		}
		b.WriteString(fmt.Sprintf("%-16s %s\n", summaryLabel(label), value))
	}

	return b.String()
}

func summaryLabel(label string) string {
	if label == "orm" {
		return "ORM"
	}

	words := strings.Fields(strings.ReplaceAll(label, "-", " "))
	for index, word := range words {
		if word == "" {
			continue
		}
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
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
		case core.ActionKindWriteFile, core.ActionKindWriteEnv:
			target := filepath.Join(baseDir, action.Path)
			parentDir := filepath.Dir(target)
			if err := os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("create parent directory for %s: %w", action.Name, err)
			}
			if err := os.WriteFile(target, []byte(action.Content), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", action.Name, err)
			}
		case core.ActionKindNote:
			fmt.Fprintf(stdout, "   note: %s\n", action.Description)
		default:
			return fmt.Errorf("unsupported action kind %q", action.Kind)
		}
	}

	return nil
}

func shouldInstallAgentSkills(selection core.TemplateSelection) bool {
	for _, value := range selection.Multi(core.GroupFinalization) {
		if value == "install-agent-skills" {
			return true
		}
	}
	return false
}

func agentSkillsFragment(selection core.TemplateSelection, entries []registry.EntryDefinition) core.PlanFragment {
	sources := make([]string, 0)
	seen := make(map[string]struct{})
	for _, entry := range entries {
		for _, source := range entrySkillSources(entry) {
			if _, ok := seen[source]; ok {
				continue
			}
			seen[source] = struct{}{}
			sources = append(sources, source)
		}
	}
	slices.Sort(sources)

	actions := make([]core.ExecutionAction, 0, len(sources))
	for _, source := range sources {
		actions = append(actions, core.ExecutionAction{
			ID:          "install-skill-" + skillSourceSlug(source),
			Name:        "Install agent skills",
			Description: "Installs project-local agent skills for " + source + ".",
			Kind:        core.ActionKindCommand,
			Dir:         selection.ProjectName,
			Command:     "npx",
			Args:        []string{"skills", "add", source, "--agent", "codex", "--yes"},
		})
	}

	return core.PlanFragment{
		ID:      "agent-skills-finalization",
		OwnerID: "install-agent-skills",
		Phase:   core.PhasePostSetup,
		Actions: actions,
	}
}

func entrySkillSources(entry registry.EntryDefinition) []string {
	sources := make([]string, 0, 2)
	if source := strings.TrimSpace(entry.Properties["skillSource"]); source != "" {
		sources = append(sources, source)
	}
	for _, source := range strings.Split(entry.Properties["skillSources"], ",") {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		sources = append(sources, source)
	}
	return sources
}

func skillSourceSlug(source string) string {
	replacer := strings.NewReplacer("https://", "", "http://", "", "github.com/", "", "/", "-", ".", "-", "_", "-")
	return replacer.Replace(source)
}

func mergeEnvWriteActions(actions []core.ExecutionAction) []core.ExecutionAction {
	merged := make([]core.ExecutionAction, 0, len(actions))
	envIndexes := make(map[string]int)
	for _, action := range actions {
		if action.Kind != core.ActionKindWriteEnv {
			merged = append(merged, action)
			continue
		}

		if index, ok := envIndexes[action.Path]; ok {
			merged[index].Content = mergeEnvContent(merged[index].Content, action.Content)
			continue
		}

		envIndexes[action.Path] = len(merged)
		merged = append(merged, action)
	}
	return merged
}

func mergeEnvContent(existing string, next string) string {
	lines := make([]string, 0)
	seen := make(map[string]struct{})
	for _, content := range []string{existing, next} {
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
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
	case core.ActionKindWriteFile:
		return fmt.Sprintf("write file %s", action.Path)
	case core.ActionKindWriteEnv:
		return fmt.Sprintf("write env file %s", action.Path)
	default:
		return string(action.Kind)
	}
}
