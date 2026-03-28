package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/I-invincib1e/http-cli/internal/config"
)

type Command struct {
	Use     string
	Short   string
	Long    string // optional detailed description
	Run     func(args []string)
	Subs    map[string]*Command
	Aliases []string
}

func (c *Command) AddCommand(sub *Command) {
	if c.Subs == nil {
		c.Subs = make(map[string]*Command)
	}
	name := strings.Split(sub.Use, " ")[0]
	c.Subs[name] = sub
	for _, alias := range sub.Aliases {
		c.Subs[alias] = sub
	}
}

// PrintHelp prints contextual help for any command
func (c *Command) PrintHelp() {
	desc := c.Short
	if c.Long != "" {
		desc = c.Long
	}
	fmt.Fprintf(os.Stderr, "%s\n\n", desc)

	if len(c.Subs) > 0 {
		fmt.Fprintf(os.Stderr, "Commands:\n")
		visited := make(map[*Command]bool)
		for k, sub := range c.Subs {
			if visited[sub] {
				continue
			}
			visited[sub] = true
			primary := strings.Split(sub.Use, " ")[0]
			if k != primary {
				continue
			}
			aliasStr := ""
			if len(sub.Aliases) > 0 {
				aliasStr = fmt.Sprintf(" (alias: %s)", strings.Join(sub.Aliases, ", "))
			}
			fmt.Fprintf(os.Stderr, "  %-14s %s%s\n", primary, sub.Short, aliasStr)
		}
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintf(os.Stderr, "Usage:\n  %s [command] [flags]\n\n", c.Use)
	fmt.Fprintf(os.Stderr, "Flags:\n  -h, --help   Show help for this command\n")
	fmt.Fprintf(os.Stderr, "\nUse \"%s [command] --help\" for more information.\n", c.Use)
}

// collectCommandNames dynamically gathers all top-level command names
func collectCommandNames(c *Command) []string {
	var names []string
	visited := make(map[*Command]bool)
	for k, sub := range c.Subs {
		if visited[sub] {
			continue
		}
		visited[sub] = true
		primary := strings.Split(sub.Use, " ")[0]
		if k != primary {
			continue
		}
		names = append(names, primary)
	}
	return names
}

var RootCmd = &Command{
	Use:   "http-cli",
	Short: "A fast and colorful HTTP CLI tool",
	Long:  "HTTP CLI is a zero-dependency, colorful command-line HTTP client\nfor developers who want Postman-like workflows in the terminal.",
}

func Execute(args []string) {
	if err := config.LoadGlobalConfig(); err != nil {
		// Soft fail
	}

	if len(args) == 0 {
		RootCmd.Run(args)
		return
	}

	arg := args[0]
	if arg == "help" || arg == "--help" || arg == "-h" {
		RootCmd.PrintHelp()
		return
	}

	if sub, ok := RootCmd.Subs[arg]; ok {
		if len(args) > 1 {
			subArg := args[1]
			if subArg == "--help" || subArg == "-h" || subArg == "help" {
				sub.PrintHelp()
				return
			}
			if sub.Subs != nil {
				if subSub, subOk := sub.Subs[subArg]; subOk {
					if len(args) > 2 && (args[2] == "--help" || args[2] == "-h") {
						subSub.PrintHelp()
						return
					}
					subSub.Run(args[2:])
					return
				}
			}
		}
		sub.Run(args[1:])
		return
	}

	if strings.HasPrefix(arg, "-") {
		RequestSendCmd.Run(args)
		return
	}

	suggestion := findClosestCommand(arg, RootCmd.Subs)
	fmt.Fprintf(os.Stderr, "Error: unknown command %q for \"http-cli\"\n", arg)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "\nDid you mean this?\n\t%s\n", suggestion)
	}
	fmt.Fprintf(os.Stderr, "\nRun 'http-cli --help' for usage.\n")
	os.Exit(1)
}

func findClosestCommand(arg string, subs map[string]*Command) string {
	best := ""
	minDist := 3
	for name := range subs {
		dist := levenshtein(arg, name)
		if dist < minDist {
			minDist = dist
			best = name
		}
	}
	return best
}

func levenshtein(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := 0; i <= len(s); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(t); j++ {
		d[0][j] = j
	}
	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			nx := d[i-1][j] + 1
			ny := d[i][j-1] + 1
			nxy := d[i-1][j-1] + cost
			m := nx
			if ny < m {
				m = ny
			}
			if nxy < m {
				m = nxy
			}
			d[i][j] = m
		}
	}
	return d[len(s)][len(t)]
}

func init() {
	RootCmd.Run = func(args []string) {
		if len(args) > 0 && strings.HasPrefix(args[0], "-") {
			RequestSendCmd.Run(args)
			return
		}
		RootCmd.PrintHelp()
	}
}
