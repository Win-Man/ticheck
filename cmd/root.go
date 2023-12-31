package cmd

import (
	"github.com/Win-Man/ticheck/service"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

//sql-diff flags
var configPath string
var logLevel string
var logPath string
var output string
var version bool

const (
	COMPONENT_TIDB = "tidb"
	COMPONENT_PD   = "pd"
	COMPONENT_TIKV = "tikv"
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// cobra.OnInitialize(initConfig)

	rootCmd = &cobra.Command{
		Use:   "ticheck",
		Short: "ticheck command tool",
		Long:  `A command tool for TiDB check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			service.GetAppVersion(version)

			return nil
		},
	}
	rootCmd.AddCommand(newArgsCheckCmd(), newDRCheckCmd(), newDBCheckCmd(), newCDCCheckCmd())
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "view ticheck version")

}
