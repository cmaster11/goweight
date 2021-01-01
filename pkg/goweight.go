package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/mattn/go-zglob"
)

var moduleRegex = regexp.MustCompile("packagefile (.*)=(.*)")

func processModule(line string) *ModuleEntry {
	captures := moduleRegex.FindAllStringSubmatch(line, -1)
	if captures == nil {
		return nil
	}
	tmpPath := captures[0][2]
	dirName := filepath.Dir(tmpPath)
	stat, _ := os.Stat(tmpPath)
	sz := uint64(stat.Size())

	return &ModuleEntry{
		Path:      tmpPath,
		DirName:   dirName,
		Name:      captures[0][1],
		Size:      sz,
		SizeHuman: humanize.Bytes(sz),
	}
}

type ModuleEntry struct {
	Path      string `json:"path"`
	DirName   string `json:"dirName"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	SizeHuman string `json:"size_human"`
}
type GoWeight struct {
	EnvVars  map[string]string
	BuildCmd []string
	Debug    bool
}

func NewGoWeight() *GoWeight {
	return &GoWeight{
		BuildCmd: []string{"go", "build", "-o", "goweight-bin-target", "-work", "-a"},
		EnvVars:  map[string]string{},
	}
}

func (g *GoWeight) BuildCurrent() string {
	d := strings.Split(strings.TrimSpace(g.run(g.BuildCmd, g.EnvVars)), "\n")[0]
	return strings.Split(strings.TrimSpace(d), "=")[1]
}

func (g *GoWeight) run(cmd []string, envVars map[string]string) string {
	cmdToExec := exec.Command(cmd[0], cmd[1:]...)

	var envVarsArray []string
	for key, value := range envVars {
		envVarsArray = append(envVarsArray, fmt.Sprintf("%s=%s", key, value))
	}

	envOs := os.Environ()

	cmdToExec.Env = append(envOs,
		envVarsArray...,
	)

	if g.Debug {
		log.Printf("running with env:\n%v", cmdToExec.Env)
	}

	out, err := cmdToExec.CombinedOutput()
	if err != nil {
		log.Fatalf("%s\n%s", err, string(out))
	}
	os.Remove("goweight-bin-target")
	return string(out)
}

func (g *GoWeight) Process(work string) map[string][]*ModuleEntry {
	files, err := zglob.Glob(work + "**/importcfg")
	if err != nil {
		log.Fatal(err)
	}

	/*
		Process all dependencies
	*/

	filesDepsMap := make(map[string][]*ModuleEntry)
	processedPackagesMap := make(map[string]string)

	for _, file := range files {
		dirName := filepath.Dir(file)
		f, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		/*
			# import config
			packagefile errors=/tmp/go-build558891369/b005/_pkg_.a
			packagefile fmt=/tmp/go-build558891369/b031/_pkg_.a
			packagefile io=/tmp/go-build558891369/b015/_pkg_.a
			packagefile os=/tmp/go-build558891369/b034/_pkg_.a
			packagefile reflect=/tmp/go-build558891369/b029/_pkg_.a
		*/

		lines := strings.Split(string(f), "\n")
		for _, line := range lines {
			module := processModule(line)
			if module == nil {
				continue
			}
			processedPackagesMap[module.DirName] = module.Name
			filesDepsMap[dirName] = append(filesDepsMap[dirName], module)
		}
	}

	remappedModuleNamesMap := make(map[string][]*ModuleEntry)

	// Recompute module names
	for dirName, entry := range filesDepsMap {
		if moduleName, found := processedPackagesMap[dirName]; found {
			remappedModuleNamesMap[moduleName] = entry
			continue
		}

		remappedModuleNamesMap[dirName] = entry
	}

	return remappedModuleNamesMap
}
