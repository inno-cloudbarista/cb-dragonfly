package get

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Gets CB-Dragonfly configuration",
		Long:  ``,
		RunE:  getConfigRun,
	}
	return cmd
}

func getConfigRun(cmd *cobra.Command, args []string) error {
	fmt.Println("getConfigRun()!")
	return nil
}
