package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Makefile struct {
	Targets   Targets
	Variables Variables
}

// Variables contains make variables
type Variables map[string][]string

// Targets contains makefile targets
type Targets map[string]struct{}

func ParseMakefile() (Makefile, error) {
	command := exec.Command("make", "-pRrq")

	output, err := command.StdoutPipe()
	if err != nil {
		return Makefile{}, fmt.Errorf("could not get stdout pipe: %w", err)
	}

	if err := command.Start(); err != nil {
		return Makefile{}, fmt.Errorf("failed to start command 'make -pRrq': %w", err)
	}

	// Iterate over lines, building a map of all variables
	results := Makefile{Variables: make(Variables), Targets: make(Targets)}
	variableMatcher := regexp.MustCompile(`^([^:#=\s]+)\s+:=\s+(.*)$`)
	targetMatcher := regexp.MustCompile(`^([^.:\s][^:\s]+):`)
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := variableMatcher.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				results.Variables[match[1]] = spaceSeparated(match[2])
			}
		}

		if matches := targetMatcher.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				results.Targets[match[1]] = struct{}{}
			}
		}
	}

	_ = command.Wait()
	return results, nil
}

var variableMatcher = regexp.MustCompile(`\$[{(]([^:#=\s]+)[})]`)

func (v Variables) Expand(str string) []string {
	// Current expanded variables
	expanded := []string{str}

	// For each variable in the string we need to expand it
	for _, matched := range variableMatcher.FindAllStringIndex(str, -1) {

		// We loop over the current set of expanded strings, and for the current
		// variable expand it for every value we have stored.
		//
		// For example:
		//   $(a)-$(b), where a = [foo, bar] and b = [baz, qux]
		//
		// Will first expand to
		//   foo-$(b)
		//   bar-$(b)
		//
		// Then expand to
		//   foo-baz
		//   foo-qux
		//   bar-baz
		//   bar-qux

		furtherExpanded := []string{}
		for _, e := range expanded {
			// Prefix and suffix is the string surrounding the variable
			prefix := e[:matched[0]]
			suffix := e[matched[1]:]

			// We know the name starts 2 characters in and the last character is
			// the closing brace
			name := e[matched[0]+2 : matched[1]-1]

			// If we have no entries for the variable, don't expand it, otherwise
			// a new string is build from the prefix, the value and the suffix
			if len(v[name]) == 0 {
				furtherExpanded = append(furtherExpanded, e)
			} else {
				for _, value := range v[name] {
					furtherExpanded = append(furtherExpanded, prefix+value+suffix)
				}
			}
		}

		expanded = furtherExpanded
	}

	return expanded
}

func (v Variables) Replace(str string) string {
	return variableMatcher.ReplaceAllStringFunc(str, func(s string) string {
		name := variableMatcher.FindAllStringSubmatch(s, -1)[0][1]
		if len(v[name]) != 0 {
			return v[name][0]
		}

		return s
	})
}

func (t Targets) Exists(name string) bool {
	_, exists := t[name]
	return exists
}

func spaceSeparated(v string) []string {
	r := csv.NewReader(strings.NewReader(v))
	r.Comma = ' '
	s, _ := r.Read()
	return s
}
