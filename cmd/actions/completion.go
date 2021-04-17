package actions

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(squadron completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ squadron completion bash > /etc/bash_completion.d/squadron
  # macOS:
  $ squadron completion bash > /usr/local/etc/bash_completion.d/squadron

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ squadron completion zsh > "${fpath[1]}/_squadron"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ squadron completion fish | source

  # To load completions for each session, execute once:
  $ squadron completion fish > ~/.config/fish/completions/squadron.fish

PowerShell:

  PS> squadron completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> squadron completion powershell > squadron.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
