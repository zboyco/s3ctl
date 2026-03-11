package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var noDesc bool

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate shell completion scripts",
	Long:                  "Generate shell completion scripts for bash, zsh, fish, or powershell.",
	Args:                  cobra.ExactValidArgs(1),
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	DisableFlagsInUseLine: true,
	Example: `  s3ctl completion bash > /etc/bash_completion.d/s3ctl
  s3ctl completion zsh > "${fpath[1]}/_s3ctl"
  s3ctl completion fish > ~/.config/fish/completions/s3ctl.fish
  s3ctl completion powershell > s3ctl.ps1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, !noDesc)
		case "zsh":
			if noDesc {
				return cmd.Root().GenZshCompletionNoDesc(os.Stdout)
			}
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, !noDesc)
		case "powershell":
			if noDesc {
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

func init() {
	completionCmd.Flags().BoolVar(&noDesc, "no-descriptions", false, "disable completion descriptions")
	rootCmd.AddCommand(completionCmd)
}
