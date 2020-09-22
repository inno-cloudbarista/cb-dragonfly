package cbmon

import (
	"github.com/spf13/cobra"

	"github.com/cloud-barista/cb-dragonfly/pkg/cmd/cbmon/cmd/get"
	"github.com/cloud-barista/cb-dragonfly/pkg/cmd/cbmon/cmd/version"
)

// GetCLIRoot returns root command for CB-MON
func GetCLIRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "cbmon",
		Short: "CB-MON Command Line Interface for Cloud-Barista CB-Dragonfly framework",
	}
	root.AddCommand(
		version.NewCmd(),
		get.NewCmd(),
	)
	return root
}
