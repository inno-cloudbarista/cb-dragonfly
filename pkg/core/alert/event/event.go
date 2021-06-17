package event

import (
	"encoding/json"
	"strings"

	"github.com/cloud-barista/cb-dragonfly/pkg/cbstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/types"
)

func CreateEventLog(eventLog types.AlertEventLog) error {
	// 태스크 로그 목록 업데이트
	err := saveEventLog(eventLog.Id, eventLog)
	if err != nil {
		return err
	}
	// 전체 태스크 로그 목록 업데이트
	err = saveEventLog("all", eventLog)
	if err != nil {
		return err
	}
	return nil
}

func ListEventLog(taskId string, logLevel string) ([]types.AlertEventLog, error) {
	var eventLogArr []types.AlertEventLog

	if taskId == "" {
		taskId = "all"
	}
	eventLogStr := cbstore.GetInstance().StoreGet(taskId)
	if eventLogStr == "" {
		return []types.AlertEventLog{}, nil
	}
	err := json.Unmarshal([]byte(eventLogStr), &eventLogArr)
	if err != nil {
		return nil, err
	}
	if logLevel == "" {
		return eventLogArr, nil
	}

	filterdEventLogArr := []types.AlertEventLog{}
	for _, log := range eventLogArr {
		if strings.EqualFold(log.Level, logLevel) {
			filterdEventLogArr = append(filterdEventLogArr, log)
		}
	}
	return filterdEventLogArr, nil
}

func DeleteEventLog(taskId string) error {
	return cbstore.GetInstance().StoreDelete(taskId)
}

func saveEventLog(key string, eventLog types.AlertEventLog) error {
	var eventLogArr []types.AlertEventLog
	eventLogStr := cbstore.GetInstance().StoreGet(key)

	if eventLogStr != "" {
		// Get event log array
		err := json.Unmarshal([]byte(eventLogStr), &eventLogArr)
		if err != nil {
			return err
		}
	}

	// Add new event log
	eventLogArr = append(eventLogArr, eventLog)

	// Save event log
	newEventLogBytes, err := json.Marshal(eventLogArr)
	if err != nil {
		return err
	}
	err = cbstore.GetInstance().StorePut(key, string(newEventLogBytes))
	if err != nil {
		return err
	}
	return nil
}
