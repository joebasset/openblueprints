package main

import (
	"fmt"
	"os"

	"openblueprints/internals/cli"
	"openblueprints/internals/planner"
	"openblueprints/internals/registry"
	"openblueprints/internals/resolver"
	"openblueprints/internals/tui"
)

func main() {
	reg := registry.New()
	planBuilder := planner.New(reg)
	planResolver := resolver.New(reg, planBuilder)

	app := cli.New(planResolver, planBuilder, tui.New(planResolver, planBuilder))
	if err := app.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
