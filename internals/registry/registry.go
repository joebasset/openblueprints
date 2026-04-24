package registry

import (
	"fmt"

	"openblueprints/internals/core"
)

type GroupDefinition struct {
	ID             core.ChoiceGroupID
	Title          string
	Description    string
	Required       bool
	Multi          bool
	ActivationCaps []core.Capability
}

type EntryDefinition struct {
	ID          string
	Name        string
	Group       core.ChoiceGroupID
	IsDefault   bool
	Provides    []core.Capability
	RequiresAll []core.Capability
	Excludes    []core.Capability
	Fragments   []core.FragmentBuilder
	Properties  map[string]string
}

type Registry struct {
	groups         []GroupDefinition
	groupByID      map[core.ChoiceGroupID]GroupDefinition
	entriesByID    map[core.ChoiceGroupID]map[string]EntryDefinition
	entriesByGroup map[core.ChoiceGroupID][]EntryDefinition
}

func New() Registry {
	r := Registry{
		groupByID:      make(map[core.ChoiceGroupID]GroupDefinition),
		entriesByID:    make(map[core.ChoiceGroupID]map[string]EntryDefinition),
		entriesByGroup: make(map[core.ChoiceGroupID][]EntryDefinition),
	}

	registerGroups(&r)
	registerNextFrontend(&r)
	registerExpoFrontend(&r)
	registerTanStackStartFrontend(&r)
	registerBackendExpress(&r)
	registerBackendHono(&r)
	registerBackendHonoCloudflareWorkers(&r)
	registerBackendNext(&r)
	registerBackendGo(&r)
	registerDatabases(&r)
	registerORMs(&r)
	registerPackageManagers(&r)
	registerCodeQuality(&r)
	registerAuthAddons(&r)
	registerEmailAddons(&r)
	registerCloudflareAddons(&r)
	registerStorageAddons(&r)
	registerFinalization(&r)

	return r
}

func (r *Registry) RegisterGroup(group GroupDefinition) {
	r.groups = append(r.groups, group)
	r.groupByID[group.ID] = group
}

func (r *Registry) RegisterEntry(entry EntryDefinition) {
	if r.entriesByID[entry.Group] == nil {
		r.entriesByID[entry.Group] = make(map[string]EntryDefinition)
	}
	r.entriesByID[entry.Group][entry.ID] = entry
	r.entriesByGroup[entry.Group] = append(r.entriesByGroup[entry.Group], entry)
}

func (r Registry) OrderedGroups() []GroupDefinition {
	return append([]GroupDefinition(nil), r.groups...)
}

func (r Registry) EntriesForGroup(groupID core.ChoiceGroupID) []EntryDefinition {
	return append([]EntryDefinition(nil), r.entriesByGroup[groupID]...)
}

func (r Registry) Entry(groupID core.ChoiceGroupID, entryID string) (EntryDefinition, bool) {
	entry, ok := r.entriesByID[groupID][entryID]
	return entry, ok
}

func (r Registry) SelectedEntries(selection core.TemplateSelection) ([]EntryDefinition, error) {
	selected := make([]EntryDefinition, 0)
	for _, group := range r.groups {
		if group.Multi {
			for _, entryID := range selection.Multi(group.ID) {
				entry, ok := r.Entry(group.ID, entryID)
				if !ok {
					return nil, fmt.Errorf("unknown entry %q", entryID)
				}
				selected = append(selected, entry)
			}
			continue
		}

		entryID := selection.Single(group.ID)
		if entryID == "" {
			continue
		}
		entry, ok := r.Entry(group.ID, entryID)
		if !ok {
			return nil, fmt.Errorf("unknown entry %q", entryID)
		}
		selected = append(selected, entry)
	}
	return selected, nil
}
