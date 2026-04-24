package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerBackendGo(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "go-api",
		Name:  "Go API",
		Group: core.GroupBackend,
		Provides: []core.Capability{
			"backend:selected",
			"backend:go-api",
			"backend:go",
			"workspace:backend",
		},
		RequiresAll: []core.Capability{"frontend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				dir := backendDir(selection)
				return []core.PlanFragment{
					{
						ID:      "go-backend",
						OwnerID: "go-api",
						Phase:   core.PhaseScaffold,
						Actions: []core.ExecutionAction{
							{
								ID:          "go-dir",
								Name:        "Create backend workspace",
								Description: "Creates the Go backend directory.",
								Kind:        core.ActionKindMkdir,
								Path:        dir,
							},
							commandAction("go-init", "Initialize Go module", "Creates a Go module for the backend workspace.", dir, "go", "mod", "init", "backend"),
						},
					},
					{
						ID:      "go-backend-files",
						OwnerID: "go-api",
						Phase:   core.PhaseIntegration,
						Actions: []core.ExecutionAction{
							writeFileAction("go-main", "Write Go API entrypoint", "Adds the initial Go HTTP server entrypoint.", filepath.Join(backendDir(selection), "main.go"), goMainSource()),
						},
					},
				}
			},
		},
		Properties: map[string]string{"kind": "pack"},
	})
}

func goMainSource() string {
	return `package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	http.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte("{\"status\":\"ok\"}"))
	})

	log.Printf("go api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}`
}
