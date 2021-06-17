package alert

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/task"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/event"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/alert/types"
)

func CreateEventLog(c echo.Context) error {
	var jsonMap map[string]interface{}
	if err := c.Bind(&jsonMap); err != nil {
		return err
	}

	var eventLog types.AlertEventLog
	err := mapstructure.Decode(jsonMap, &eventLog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = event.CreateEventLog(eventLog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusNoContent, "")
}

func ListEventLog(c echo.Context) error {
	taskName := c.Param("task_id")
	logLevel := c.QueryParam("level")
	if taskName != "" {
		taskName = fmt.Sprintf(task.KapacitorTaskFormat, taskName)
	}
	alertLogList, err := event.ListEventLog(taskName, logLevel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, alertLogList)
}
