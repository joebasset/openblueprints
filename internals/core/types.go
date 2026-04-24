package core

import (
	"fmt"
	"slices"
	"strings"
)

type ChoiceGroupID string

const (
	GroupFrontend       ChoiceGroupID = "frontend"
	GroupBackend        ChoiceGroupID = "backend"
	GroupDatabase       ChoiceGroupID = "database"
	GroupORM            ChoiceGroupID = "orm"
	GroupPackageManager ChoiceGroupID = "packageManager"
	GroupCodeQuality    ChoiceGroupID = "codeQuality"
	GroupAddon          ChoiceGroupID = "addon"
	GroupStorage        ChoiceGroupID = "storage"
	GroupFinalization   ChoiceGroupID = "finalization"
)

var OrderedChoiceGroups = []ChoiceGroupID{
	GroupFrontend,
	GroupBackend,
	GroupDatabase,
	GroupStorage,
	GroupORM,
	GroupPackageManager,
	GroupCodeQuality,
	GroupAddon,
	GroupFinalization,
}

var choiceGroupLabels = map[ChoiceGroupID]string{
	GroupFrontend:       "frontend",
	GroupBackend:        "backend",
	GroupDatabase:       "database",
	GroupORM:            "orm",
	GroupPackageManager: "package manager",
	GroupCodeQuality:    "linting / formatter",
	GroupAddon:          "addons",
	GroupStorage:        "storage",
	GroupFinalization:   "finalization",
}

type Capability string

type FragmentBuilder func(TemplateSelection) []PlanFragment

type ChoiceGroup struct {
	ID          ChoiceGroupID
	Title       string
	Description string
	Required    bool
	Multi       bool
	Choices     []ResolvedChoice
}

type ResolvedChoice struct {
	ID         string
	Name       string
	Available  bool
	IsDefault  bool
	Properties map[string]string
}

type TemplateSelection struct {
	ProjectName string
	Singles     map[ChoiceGroupID]string
	Multis      map[ChoiceGroupID][]string
	Touched     map[ChoiceGroupID]bool
}

func NewTemplateSelection() TemplateSelection {
	return TemplateSelection{
		Singles: make(map[ChoiceGroupID]string),
		Multis:  make(map[ChoiceGroupID][]string),
		Touched: make(map[ChoiceGroupID]bool),
	}
}

func (s TemplateSelection) Clone() TemplateSelection {
	cloned := NewTemplateSelection()
	cloned.ProjectName = s.ProjectName
	for key, value := range s.Singles {
		cloned.Singles[key] = value
	}
	for key, value := range s.Multis {
		cloned.Multis[key] = slices.Clone(value)
	}
	for key, value := range s.Touched {
		cloned.Touched[key] = value
	}
	return cloned
}

func (s *TemplateSelection) Normalize() {
	for key, values := range s.Multis {
		s.Multis[key] = uniqueNonEmpty(values)
	}
}

func (s TemplateSelection) Single(groupID ChoiceGroupID) string {
	if s.Singles == nil {
		return ""
	}
	return s.Singles[groupID]
}

func (s *TemplateSelection) SetSingle(groupID ChoiceGroupID, value string) {
	if s.Singles == nil {
		s.Singles = make(map[ChoiceGroupID]string)
	}
	if s.Touched == nil {
		s.Touched = make(map[ChoiceGroupID]bool)
	}
	s.Singles[groupID] = value
	s.Touched[groupID] = true
}

func (s TemplateSelection) Multi(groupID ChoiceGroupID) []string {
	if s.Multis == nil {
		return nil
	}
	return slices.Clone(s.Multis[groupID])
}

func (s *TemplateSelection) SetMulti(groupID ChoiceGroupID, values []string) {
	if s.Multis == nil {
		s.Multis = make(map[ChoiceGroupID][]string)
	}
	if s.Touched == nil {
		s.Touched = make(map[ChoiceGroupID]bool)
	}
	s.Multis[groupID] = uniqueNonEmpty(values)
	s.Touched[groupID] = true
}

func (s TemplateSelection) IsTouched(groupID ChoiceGroupID) bool {
	if s.Touched == nil {
		return false
	}
	return s.Touched[groupID]
}

func (s TemplateSelection) HasSelection(groupID ChoiceGroupID) bool {
	if groupID == GroupAddon || groupID == GroupStorage || groupID == GroupFinalization {
		return s.IsTouched(groupID)
	}
	return s.Single(groupID) != ""
}

func (s TemplateSelection) SummaryLines() []string {
	lines := []string{fmt.Sprintf("project: %s", s.ProjectName)}
	for _, groupID := range OrderedChoiceGroups {
		label := choiceGroupLabels[groupID]
		if groupID == GroupAddon || groupID == GroupStorage || groupID == GroupFinalization {
			values := s.Multi(groupID)
			if len(values) > 0 {
				lines = append(lines, fmt.Sprintf("%s: %s", label, strings.Join(values, ", ")))
			}
			continue
		}

		value := s.Single(groupID)
		if value != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", label, value))
		}
	}
	return lines
}

func ChoiceGroupLabel(groupID ChoiceGroupID) string {
	if label, ok := choiceGroupLabels[groupID]; ok {
		return label
	}
	return string(groupID)
}

type ActionKind string

const (
	ActionKindMkdir     ActionKind = "mkdir"
	ActionKindCommand   ActionKind = "command"
	ActionKindNote      ActionKind = "note"
	ActionKindWriteFile ActionKind = "write-file"
	ActionKindWriteEnv  ActionKind = "write-env"
)

type PlanPhase string

const (
	PhaseWorkspace    PlanPhase = "workspace"
	PhaseScaffold     PlanPhase = "scaffold"
	PhaseDependencies PlanPhase = "dependencies"
	PhaseIntegration  PlanPhase = "integration"
	PhasePostSetup    PlanPhase = "post-setup"
)

var OrderedPlanPhases = []PlanPhase{
	PhaseWorkspace,
	PhaseScaffold,
	PhaseDependencies,
	PhaseIntegration,
	PhasePostSetup,
}

type ExecutionAction struct {
	ID          string
	Name        string
	Description string
	Kind        ActionKind
	Dir         string
	Path        string
	Command     string
	Args        []string
	Content     string
}

type PlanFragment struct {
	ID      string
	OwnerID string
	Phase   PlanPhase
	Actions []ExecutionAction
}

type TemplatePlan struct {
	Selection TemplateSelection
	Fragments []PlanFragment
	Actions   []ExecutionAction
}

type Resolver interface {
	ResolveNext(selection TemplateSelection) (*ChoiceGroup, error)
	ResolveFinal(selection TemplateSelection) (TemplatePlan, error)
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		filtered = append(filtered, value)
	}
	slices.Sort(filtered)
	return filtered
}
