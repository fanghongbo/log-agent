package g

import (
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		return
	},
}

func initCmd() error {
	var (
		version    bool
		printUsage bool
		err        error
	)

	RootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", `Location of config files (example "./config/cfg.json.example")`)
	RootCmd.Flags().BoolVarP(&version, "version", "v", false, "Print version information and quit")
	RootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	RootCmd.PersistentFlags().BoolVarP(&printUsage, "help", "h", false, "Print usage")

	if err = RootCmd.Execute(); err != nil {
		os.Exit(128)
	}

	if printUsage {
		os.Exit(0)
	}

	return nil
}
