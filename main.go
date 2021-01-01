package main

import (
	"encoding/json"
	"fmt"

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

	if *jsonOutput {
		m, _ := json.Marshal(modules)
		fmt.Print(string(m))
	} else {
		for module, deps := range modules {
			fmt.Printf("%s\n", module)
			for _, dep := range deps {
				if dep == nil {
					continue
				}
				fmt.Printf("\t%8s %s\n", dep.SizeHuman, dep.Name)
			}
		}
	}
}
