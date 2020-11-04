package request

import (
	"context"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/common"
	pb "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf/cbdragonfly"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
)

type MonitoringRequest struct {
	Client  pb.MONClient
	Timeout time.Duration
	InType  string
	OutType string
}

func NewMonitoringRequest(monClient *pb.MONClient, timeout time.Duration, inType string, outType string) *MonitoringRequest {
	newMonReq := MonitoringRequest{
		Client:  *monClient,
		Timeout: timeout,
		InType:  inType,
		OutType: outType,
	}
	return &newMonReq
}

/* Monitoring Configuration API */

// GetMonitoringConfig
func (monReq *MonitoringRequest) GetMonitoringConfig() (string, error) {
	// set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), monReq.Timeout)
	defer cancel()

	resp, err := monReq.Client.GetMonConfig(ctx, &pb.Empty{})
	if err != nil {
		return "", err
	}
	return monReq.convertResponseToString(resp)
}

// SetMonitoringConfig
func (monReq *MonitoringRequest) SetMonitoringConfig(config pb.MonitoringConfigInfo) (string, error) {
	// set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), monReq.Timeout)
	defer cancel()

	reqParams := &pb.MonitoringConfigRequest{Item: &config}
	resp, err := monReq.Client.SetMonConfig(ctx, reqParams)
	if err != nil {
		return "", err
	}
	return monReq.convertResponseToString(resp)
}

// ResetMonitoringConfig
func (monReq *MonitoringRequest) ResetMonitoringConfig() (string, error) {
	// set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), monReq.Timeout)
	defer cancel()

	resp, err := monReq.Client.ResetMonConfig(ctx, &pb.Empty{})
	if err != nil {
		return "", err
	}
	return monReq.convertResponseToString(resp)
}

/* VM Monitoring API */

// GetVMMonInfo
func (monReq *MonitoringRequest) GetVMMonInfo(metricName string, vmMonQueryRequest pb.VMMonQryRequest) (string, error) {
	// set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), monReq.Timeout)
	defer cancel()

	var resp interface{}
	var err error

	switch metricName {
	case metric.Cpu:
		resp, err = monReq.Client.GetVMMonCpuInfo(ctx, &vmMonQueryRequest)
	case metric.CpuFreqency:
		resp, err = monReq.Client.GetVMMonCpuFreqInfo(ctx, &vmMonQueryRequest)
	case metric.Memory:
		resp, err = monReq.Client.GetVMMonMemoryInfo(ctx, &vmMonQueryRequest)
	case metric.Disk:
		resp, err = monReq.Client.GetVMMonDiskInfo(ctx, &vmMonQueryRequest)
	case metric.DiskIO:
		resp, err = monReq.Client.GetVMMonCpuFreqInfo(ctx, &vmMonQueryRequest)
	case metric.Network:
		resp, err = monReq.Client.GetVMMonCpuFreqInfo(ctx, &vmMonQueryRequest)
	}

	if err != nil {
		return "", err
	}
	return monReq.convertResponseToString(resp)
}

// GetVMOnDemandMonInfo
func (monReq *MonitoringRequest) GetVMOnDemandMonInfo(metricName string, vmMonQueryRequest pb.VMOndemandMonQryRequest) (string, error) {
	// set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), monReq.Timeout)
	defer cancel()

	var resp interface{}
	var err error

	switch metricName {
	case metric.Cpu:
		resp, err = monReq.Client.GetVMOndemandMonCpuInfo(ctx, &vmMonQueryRequest)
	case metric.CpuFreqency:
		resp, err = monReq.Client.GetVMOndemandMonCpuFreqInfo(ctx, &vmMonQueryRequest)
	}

	if err != nil {
		return "", err
	}
	return monReq.convertResponseToString(resp)
}

// convertResponseToString - convert response object to string
func (monReq *MonitoringRequest) convertResponseToString(response interface{}) (string, error) {
	result, err := common.ConvertToOutput(monReq.OutType, response)
	if err != nil {
		return "", err
	}
	return result, nil
}
