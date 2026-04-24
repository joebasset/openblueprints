package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerCloudflareAddons(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "cloudflare-kv",
		Name:  "Cloudflare KV",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:kv",
			"provider:cloudflare-kv",
		},
		RequiresAll: []core.Capability{
			"runtime:cloudflare-workers",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "cloudflare-kv-addon",
					OwnerID: "cloudflare-kv",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						writeFileAction("cloudflare-kv-helper", "Write Cloudflare KV helper", "Adds a typed helper for the Worker KV binding.", filepath.Join(backendDir(selection), "src", "cloudflare", "kv.ts"), cloudflareKVSource()),
					},
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "https://github.com/cloudflare/skills",
		},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "cloudflare-queues",
		Name:  "Cloudflare Queues",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:queue",
			"provider:cloudflare-queues",
		},
		RequiresAll: []core.Capability{
			"runtime:cloudflare-workers",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return []core.PlanFragment{{
					ID:      "cloudflare-queues-addon",
					OwnerID: "cloudflare-queues",
					Phase:   core.PhaseIntegration,
					Actions: []core.ExecutionAction{
						writeFileAction("cloudflare-queue-helper", "Write Cloudflare Queue helper", "Adds a typed helper for the Worker Queue binding.", filepath.Join(backendDir(selection), "src", "cloudflare", "queue.ts"), cloudflareQueueSource()),
					},
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "https://github.com/cloudflare/skills",
		},
	})
}

func cloudflareKVSource() string {
	return `export async function readJSON<T>(namespace: KVNamespace, key: string) {
  const value = await namespace.get<T>(key, "json");
  console.log("cloudflare kv read", { key, found: value !== null });
  return value;
}

export async function writeJSON<T>(namespace: KVNamespace, key: string, value: T) {
  await namespace.put(key, JSON.stringify(value));
  console.log("cloudflare kv write", { key });
}`
}

func cloudflareQueueSource() string {
	return `export async function enqueueJob<T>(queue: Queue, body: T) {
  await queue.send(body);
  console.log("cloudflare queue job enqueued");
}`
}
