package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for lu.

When installing lu through a package manager, completions may already be
configured automatically. For Homebrew, see https://docs.brew.sh/Shell-Completion

If you need to set up completions manually, follow the instructions below.

### bash

First, ensure that you have bash-completion installed:

  $ brew install bash-completion   # macOS
  $ apt install bash-completion    # Debian/Ubuntu

Then add this to your ~/.bash_profile:

  eval "$(lu completion bash)"

Or install the script system-wide:

  $ lu completion bash > /etc/bash_completion.d/lu

### zsh

Generate a _lu completion script and put it in your $fpath:

  $ mkdir -p ~/.zsh/completions
  $ lu completion zsh > ~/.zsh/completions/_lu

Ensure the following is present in your ~/.zshrc:

  fpath=(~/.zsh/completions $fpath)
  autoload -U compinit
  compinit -i

Then reload your shell:

  $ exec zsh

### fish

Generate a lu.fish completion script:

  $ lu completion fish > ~/.config/fish/completions/lu.fish

Then reload your shell:

  $ exec fish
`,
		Example: `  lu completion bash        # output bash completion script
  lu completion zsh         # output zsh completion script
  lu completion fish        # output fish completion script
  lu completion bash > /etc/bash_completion.d/lu  # install for bash`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %q (supported: bash, zsh, fish)", args[0])
			}
		},
	}

	return completionCmd
}
