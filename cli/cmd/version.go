package cmd

import (
	"github.com/spf13/cobra"
)

var versionLove bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print IFLy version",
	RunE: func(cmd *cobra.Command, args []string) error {
		tagline := "I'm Feeling Lucky"
		if versionLove {
			tagline = "I Fucking Love You \U0001F49C"
		}
		cmd.Printf("IFLy v%s — %s\n", cliVersion, tagline)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionLove, "love", false, "")
	_ = versionCmd.Flags().MarkHidden("love")
	rootCmd.AddCommand(versionCmd)
}
