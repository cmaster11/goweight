package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jondot/goweight/pkg"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	jsonOutput = kingpin.Flag("json", "Output json").Short('j').Bool()
	buildTags  = kingpin.Flag("tags", "Build tags").String()
	envVars    = kingpin.Flag("env", "Environment variables to pass").StringMap()
	debug      = kingpin.Flag("debug", "Turns on more debugging output").Bool()
	packages   = kingpin.Arg("packages", "Packages to build").String()
)

type Result struct {
	Summary       []*pkg.ModuleEntry
	DependencyMap map[string][]*pkg.ModuleEntry
}

func main() {
	kingpin.Version(fmt.Sprintf("%s (%s)", version, commit))
	kingpin.Parse()
	weight := pkg.NewGoWeight()
	if *buildTags != "" {
		weight.BuildCmd = append(weight.BuildCmd, "-tags", *buildTags)
	}
	if *packages != "" {
		weight.BuildCmd = append(weight.BuildCmd, *packages)
	}
	weight.EnvVars = *envVars
	weight.Debug = *debug

	work := weight.BuildCurrent()
	modules := weight.Process(work)

	result := &Result{
		DependencyMap: modules,
	}

	// Compute summary
	{
		allModulesMap := make(map[string]*pkg.ModuleEntry)

		for _, moduleEntries := range modules {
			for _, moduleEntry := range moduleEntries {
				allModulesMap[moduleEntry.Name] = moduleEntry
			}
		}

		var allModules []*pkg.ModuleEntry

		for _, entry := range allModulesMap {
			allModules = append(allModules, entry)
		}

		sort.Slice(allModules, func(i, j int) bool { return allModules[i].Size > allModules[j].Size })

		result.Summary = allModules
	}

	if *jsonOutput {
		m, _ := json.Marshal(result)
		fmt.Print(string(m))
	} else {
		for module, deps := range result.DependencyMap {
			fmt.Printf("%s\n", module)
			for _, dep := range deps {
				if dep == nil {
					continue
				}
				fmt.Printf("\t%8s %s\n", dep.SizeHuman, dep.Name)
			}
		}

		fmt.Printf("\n")

		for _, module := range result.Summary {
			fmt.Printf("%8s %s\n", module.SizeHuman, module.Name)
		}
	}
}
