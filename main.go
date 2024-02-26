package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const header = `Usage: make [targets]

Make use used to build, develop and test this project. The make targets are a mix of project specific targets, and shared targets sourced from https://github.com/cert-manager/makefile-modules. 

For more information on make see https://www.gnu.org/software/make/manual/make.html

Available targets are listed below:

`

func main() {
	makefile, err := ParseMakefile()
	if err != nil {
		panic(err)
	}

	targets := map[string]Target{}
	for _, file := range makefile.Variables["MAKEFILE_LIST"] {
		if err := extractHelpTargets(file, targets, makefile); err != nil {
			panic(err)
		}
	}

	// Extract groups
	groups := map[string][]Target{}
	for _, target := range targets {
		groups[target.Group] = append(groups[target.Group], target)
	}

	// Get sorted list of group names
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}

	sort.Strings(groupNames)

	// Print the header
	fmt.Print(header)

	// Loop over the groups, printing each target within it.
	w := NewColorTableWriter(os.Stdout, 4)
	for _, group := range groupNames {
		// Get a sorted slice of all the targets for a consistent output
		targets := groups[group]
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Name < targets[j].Name
		})

		w.AddCell(ColorLightBlue, group)
		for _, target := range targets {
			w.AddCell(ColorGreen, target.Name)
			w.AddCell(ColorNone, makefile.Variables.Replace(target.Usage))
			w.FlushRow()
		}

		// Empty Row between groups
		w.FlushRow()
	}

	w.FlushTable()
}

type Target struct {
	Name  string
	Group string
	Usage string
}

func extractHelpTargets(path string, targets map[string]Target, makefile Makefile) error {
	// Open the makefile
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	// Scan over each line
	targetMatcher := regexp.MustCompile(`^([^.:\s][^:\s]+):`)
	target := Target{}
	scanner := bufio.NewScanner(file)
	usage := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "## @category"):
			target.Group = strings.TrimSpace(
				strings.TrimPrefix(line, "## @category"),
			)
		case strings.HasPrefix(line, "##"):
			usage = append(usage,
				strings.TrimPrefix(line, "##"),
			)
		case strings.HasPrefix(line, ".PHONY"):
			continue
		case targetMatcher.MatchString(line):
			if len(usage) != 0 {
				trimLeadingSpaces(usage)
				targetName := targetMatcher.FindAllStringSubmatch(line, -1)[0][1]
				for _, expandedName := range makefile.Variables.Expand(targetName) {
					if makefile.Targets.Exists(expandedName) {
						target.Name = expandedName
						target.Usage = strings.Join(usage, "\n")
						targets[target.Name] = target
					}
				}
			}
			fallthrough
		default:
			target = Target{}
			usage = usage[:0]
		}
	}

	return nil
}
