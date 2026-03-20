package resolver

import (
	"fmt"
	"slices"

	"openblueprints/internals/core"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
)

type Service struct {
	registry registry.Registry
	planner  planner.Service
}

func New(reg registry.Registry, planBuilder planner.Service) Service {
	return Service{
		registry: reg,
		planner:  planBuilder,
	}
}

func (s Service) ResolveNext(selection core.TemplateSelection) (*core.ChoiceGroup, error) {
	normalized := selection.Clone()
	normalized.Normalize()

	if err := s.validatePartial(normalized); err != nil {
		return nil, err
	}

	for _, group := range s.registry.OrderedGroups() {
		if !s.groupIsActive(group, normalized) {
			continue
		}
		if normalized.HasSelection(group.ID) {
			continue
		}

		choices := s.availableChoices(group.ID, normalized)
		if len(choices) == 0 {
			if group.Required {
				return nil, fmt.Errorf("no valid %s options remain for the current selection", core.ChoiceGroupLabel(group.ID))
			}
			continue
		}

		return &core.ChoiceGroup{
			ID:          group.ID,
			Title:       group.Title,
			Description: group.Description,
			Required:    group.Required,
			Multi:       group.Multi,
			Choices:     choices,
		}, nil
	}

	return nil, nil
}

func (s Service) ResolveFinal(selection core.TemplateSelection) (core.TemplatePlan, error) {
	normalized := selection.Clone()
	normalized.Normalize()

	if normalized.ProjectName == "" {
		return core.TemplatePlan{}, fmt.Errorf("project name is required")
	}
	if err := s.validatePartial(normalized); err != nil {
		return core.TemplatePlan{}, err
	}

	for _, group := range s.registry.OrderedGroups() {
		if !s.groupIsActive(group, normalized) {
			continue
		}
		if group.Required && !normalized.HasSelection(group.ID) {
			return core.TemplatePlan{}, fmt.Errorf("%s is required", core.ChoiceGroupLabel(group.ID))
		}
	}

	return s.planner.Build(normalized)
}

func (s Service) validatePartial(selection core.TemplateSelection) error {
	for _, group := range s.registry.OrderedGroups() {
		if !s.groupIsActive(group, selection) {
			if selection.HasSelection(group.ID) {
				return fmt.Errorf("%s is not valid for the current selection", core.ChoiceGroupLabel(group.ID))
			}
			continue
		}

		if !selection.HasSelection(group.ID) {
			continue
		}

		choices := s.availableChoices(group.ID, selection)
		if group.Multi {
			for _, value := range selection.Multi(group.ID) {
				if !containsChoice(choices, value) {
					return fmt.Errorf("addon %q is not valid for the current selection", value)
				}
			}
			continue
		}

		value := selection.Single(group.ID)
		if !containsChoice(choices, value) {
			return fmt.Errorf("%s %q is not valid for the current selection", core.ChoiceGroupLabel(group.ID), value)
		}
	}
	return nil
}

func (s Service) groupIsActive(group registry.GroupDefinition, selection core.TemplateSelection) bool {
	caps := s.capabilities(selection)
	for _, capability := range group.ActivationCaps {
		if !slices.Contains(caps, capability) {
			return false
		}
	}

	if !group.Required && !group.Multi {
		choices := s.availableChoices(group.ID, selection)
		return len(choices) > 0
	}

	if group.Multi {
		return len(s.availableChoices(group.ID, selection)) > 0 || selection.IsTouched(group.ID)
	}

	return true
}

func (s Service) availableChoices(groupID core.ChoiceGroupID, selection core.TemplateSelection) []core.ResolvedChoice {
	caps := s.capabilities(selection)
	choices := make([]core.ResolvedChoice, 0)
	for _, entry := range s.registry.EntriesForGroup(groupID) {
		if !meetsAll(entry.RequiresAll, caps) {
			continue
		}
		if meetsAny(entry.Excludes, caps) {
			continue
		}
		choices = append(choices, core.ResolvedChoice{
			ID:         entry.ID,
			Name:       entry.Name,
			Available:  true,
			IsDefault:  entry.IsDefault,
			Properties: cloneProperties(entry.Properties),
		})
	}
	return choices
}

func (s Service) capabilities(selection core.TemplateSelection) []core.Capability {
	capabilities := make([]core.Capability, 0)
	selected, err := s.registry.SelectedEntries(selection)
	if err != nil {
		return capabilities
	}
	for _, entry := range selected {
		capabilities = append(capabilities, entry.Provides...)
	}
	return capabilities
}

func meetsAll(required, capabilities []core.Capability) bool {
	for _, capability := range required {
		if !slices.Contains(capabilities, capability) {
			return false
		}
	}
	return true
}

func meetsAny(required, capabilities []core.Capability) bool {
	for _, capability := range required {
		if slices.Contains(capabilities, capability) {
			return true
		}
	}
	return false
}

func containsChoice(choices []core.ResolvedChoice, id string) bool {
	for _, choice := range choices {
		if choice.ID == id {
			return true
		}
	}
	return false
}

func cloneProperties(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
