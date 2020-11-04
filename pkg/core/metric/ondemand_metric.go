package metric

/*package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"

	"github.com/labstack/echo/v4"
)

// 멀티클라우드 인프라 VM 온디멘드 모니터링
func OndemandMetric(c echo.Context) error {
	//온디멘드 모니터링 Agent IP 파라미터 추출
	publicIP := c.Param("agent_ip")

	// Query Agent IP 값 체크
	if publicIP == "" {
		err := errors.New("no Agent IP in API")
		return c.JSON(http.StatusInternalServerError, err)
	}

	//Query 매트릭 값 체크
	var metricKey string
	metricName := c.Param("metric_name")
	if metricName == "" {
		//err := errors.New("no Metric Type in API")
		return c.JSON(http.StatusInternalServerError, errors.New(fmt.Sprintf("not found metric : %s", metricName)))
	}

	//온디멘드 모니터링 매트릭 파라미터 추출
	switch metricName {
	case metric.Cpu:
		metricKey = "cpu"
	case metric.CpuFreqency:
		metricKey = "cpufreq"
	case metric.Memory:
		metricKey = "mem"
	case metric.Network:
		metricKey = "net"
	default:
		return c.JSON(http.StatusInternalServerError, errors.New(fmt.Sprintf("not found metric : %s", metricName)))
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:8080/cb-dragonfly/metric/%s", publicIP, metricKey))
	if err != nil {
		return c.String(http.StatusNotImplemented, "Server Closed")
	}
	defer resp.Body.Close()
	var data = map[string]collector.TelegrafMetric{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	resultMetric, err := MappingMonMetric(metricKey, data[metricKey])
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, resultMetric)
}

func MappingMonMetric(metricKey string, metricVal collector.TelegrafMetric) (map[string]interface{}, error) {
	metricMap := map[string]interface{}{}
	metricMap["name"] = metricVal.Name
	tagMap := map[string]interface{}{
		"nsId":   metricVal.Tags["nsId"],
		"mcisId": metricVal.Tags["mcisId"],
		"vmId":   metricVal.Tags["vmId"],
	}
	metricMap["tags"] = tagMap

	metricCols, err := realtimestore.MappingMonMetric(metricKey, metricVal.Fields)
	if err != nil {
		return nil, err
	}
	metricMap["values"] = metricCols
	metricMap["time"] = time.Now().UTC() // TODO: parsing timestamp to utc time
	return metricMap, nil
}
*/
