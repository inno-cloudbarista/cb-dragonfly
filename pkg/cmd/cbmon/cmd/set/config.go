package set

import (
	"fmt"

	"github.com/spf13/cobra"

	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/request"
)

func newSetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "set configuration",
		Long:  "",
		RunE:  setConfigRun,
	}
	cmd.Flags().Int32P("agent-interval", "", 10, "Set agent interval")
	cmd.Flags().Int32P("agent-ttl", "", 10, "Set agent TTL")
	cmd.Flags().Int32P("collector-interval", "", 10, "Set collector interval")
	cmd.Flags().Int32P("scheduler-interval", "", 10, "Set agent interval")
	cmd.Flags().Int32P("max-hostcnt", "", 10, "Set maximum host count")
	return cmd
}

func setConfigRun(cmd *cobra.Command, args []string) error {
	agentInterval, _ := cmd.Flags().GetInt32("agent-interval")
	agentTTL, _ := cmd.Flags().GetInt32("agent-ttl")
	collectorInterval, _ := cmd.Flags().GetInt32("collector-interval")
	schedulerInterval, _ := cmd.Flags().GetInt32("scheduler-interval")
	maxHostCnt, _ := cmd.Flags().GetInt32("max-hostcnt")

	reqParams := pb.MonitoringConfigInfo{
		AgentInterval:     agentInterval,
		AgentTtl:          agentTTL,
		CollectorInterval: collectorInterval,
		ScheduleInterval:  schedulerInterval,
		MaxHostCount:      maxHostCnt,
	}

	monApi := request.GetMonitoringAPI()
	result, err := monApi.SetMonitoringConfig(reqParams)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
