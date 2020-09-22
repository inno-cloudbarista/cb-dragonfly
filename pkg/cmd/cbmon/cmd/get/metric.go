package get

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetMetricCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metric",
		Short: "Get Monitoring metric information",
		Long:  ``,
		RunE:  getMetricRun,
	}
	return cmd
}

func getMetricRun(cmd *cobra.Command, args []string) error {
	fmt.Println("getMetricRun()!")
	return nil
}
