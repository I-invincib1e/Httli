package cmd

import (
	"fmt"
	"os"
	"strings"
)

var CompletionCmd = &Command{
	Use:   "completion",
	Short: "Generate shell autocompletion scripts",
	Long: `Generate shell autocompletion scripts for bash, zsh, or powershell.

To load completions:

Bash:
  source <(http-cli completion bash)

  # To load completions for each session, execute once:
  http-cli completion bash > ~/.bashrc_httpcli
  echo 'source ~/.bashrc_httpcli' >> ~/.bashrc

Zsh:
  source <(http-cli completion zsh)

PowerShell:
  http-cli completion powershell | Out-String | Invoke-Expression`,
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: http-cli completion [bash|zsh|powershell]\n")
			os.Exit(1)
		}
		shell := args[0]
		topCmds := collectCommandNames(RootCmd)
		words := strings.Join(topCmds, " ")

		switch shell {
		case "bash":
			fmt.Printf(`_http_cli_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=($(compgen -W "%s" -- "$cur"))
}
complete -F _http_cli_completions http-cli
# Add this to your shell:
# source <(http-cli completion bash)
`, words)
		case "zsh":
			fmt.Printf(`#compdef http-cli
_http_cli() {
    local -a commands
    commands=(%s)
    _describe 'command' commands
}
compdef _http_cli http-cli
# Add this to your shell:
# source <(http-cli completion zsh)
`, words)
		case "powershell":
			fmt.Printf(`Register-ArgumentCompleter -CommandName http-cli -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    @(%s) | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
# Add this to your PowerShell profile:
# http-cli completion powershell | Out-String | Invoke-Expression
`, func() string {
				var quoted []string
				for _, c := range topCmds {
					quoted = append(quoted, fmt.Sprintf("'%s'", c))
				}
				return strings.Join(quoted, ", ")
			}())
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s\nSupported: bash, zsh, powershell\n", shell)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(CompletionCmd)
}
