package manager

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/kafkabroker"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore/etcd"
)

type CollectorScheduler struct {
	CheckScaling bool
	ScalePolicy string
	ScaleCount int
	CollectorCount int
	CollectorTopicMap map[int] []string
	cm *CollectManager
}

func CalculateNumberOfCollector(topicCount int, maxHostCount int) int {
	collectorCount := topicCount/maxHostCount
	if topicCount%maxHostCount != 0 || topicCount == 0 {
		collectorCount += 1
	}
	return collectorCount
}

func MergeTopicsToOneString(topicsSlice []string) string {
	var combinedTopicString string
	for _, topic := range topicsSlice {
		combinedTopicString = fmt.Sprintf(fmt.Sprintf("%s&%s", combinedTopicString, topic))
	}
	return combinedTopicString
}

func MakeCollectorTopicMap(allTopics []string, collectorCount int, maxHostCount int) map[int] []string {

	if len(allTopics) == 0 {
		return map[int] []string{}
	}

	collectorTopicMap := map[int] []string{}
	allTopicsLen := len(allTopics)
	startIdx := 0
	endIdx := 0

	for collectorCountIdx := 0; collectorCountIdx < collectorCount; collectorCountIdx++ {
		if allTopicsLen < maxHostCount {
			endIdx = len(allTopics)
		} else {
			endIdx = (collectorCountIdx+1)*maxHostCount
		}
		collectorTopicMap[collectorCountIdx] = allTopics[startIdx:endIdx]
		startIdx = endIdx
		allTopicsLen -=  maxHostCount
	}
	return collectorTopicMap
}

func ReturnDiffTopicList(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func GetAllTopicBySort(topicsSlice []string) []string {
	sort.Slice(topicsSlice, func(i, j int) bool {
		return topicsSlice[i] < topicsSlice[j]
	})
	return topicsSlice
}

func NewCollectorScheduler(manager *CollectManager) (*CollectorScheduler, error) {
	cScheduler := CollectorScheduler{
		CheckScaling:false,
		cm:manager,
		CollectorTopicMap: map[int] []string {},
	}

	return &cScheduler, nil
}

func (cScheduler CollectorScheduler) Scheduler() {

	go cScheduler.Scaler()

	beforeTopicsState := []string {}
	currentTopicsState := GetAllTopicBySort(kafkabroker.GetInstance().GetAllTopics())

	beforeMaxHostCount := config.GetInstance().Monitoring.MaxHostCount
	currentMaxHostCount := config.GetInstance().Monitoring.MaxHostCount
	for {
		fmt.Println("==== All topics ====")
		fmt.Println(currentTopicsState)
		switch {

		case currentMaxHostCount == beforeMaxHostCount && cmp.Equal(beforeTopicsState, currentTopicsState) :
			fmt.Println("case 1111")
			break

		case currentMaxHostCount != beforeMaxHostCount : // When MaxHostCount Config changed
			fmt.Println("case 2222")
			cScheduler.CollectorCount = CalculateNumberOfCollector(len(currentTopicsState), currentMaxHostCount)
			cScheduler.CollectorTopicMap = MakeCollectorTopicMap(currentTopicsState, cScheduler.CollectorCount, currentMaxHostCount)

			_ = etcd.GetInstance().DeleteMetric("topicschedule")

			for idx := 0; idx < cScheduler.CollectorCount; idx ++ {
				_ = etcd.GetInstance().WriteMetric(fmt.Sprintf("topicschedule/%d/topics", idx), MergeTopicsToOneString(cScheduler.CollectorTopicMap[idx]), false)
			}
			break

		case !cmp.Equal(beforeTopicsState, currentTopicsState):
			fmt.Println("case 3333")
			deletedTopicList := ReturnDiffTopicList(beforeTopicsState, currentTopicsState)
			fmt.Println(deletedTopicList)
			newTopicList := ReturnDiffTopicList(currentTopicsState, beforeTopicsState)
			fmt.Println(newTopicList)
			if len(cScheduler.CollectorTopicMap) == 0 {
				fmt.Println("We are here")
				cScheduler.CollectorCount = CalculateNumberOfCollector(len(currentTopicsState), currentMaxHostCount)
				topicMap := MakeCollectorTopicMap(currentTopicsState, cScheduler.CollectorCount, currentMaxHostCount)
				if len(topicMap) == 0 {
					break
				} else {
					cScheduler.CollectorTopicMap = topicMap
					_ = etcd.GetInstance().DeleteMetric("topicschedule")
					for idx := 0; idx < cScheduler.CollectorCount; idx ++ {
						_ = etcd.GetInstance().WriteMetric(fmt.Sprintf("topicschedule/%d/topics", idx), MergeTopicsToOneString(cScheduler.CollectorTopicMap[idx]), false)
					}
				}
			} else {
				leftNewTopicList := newTopicList
				for idx, topicsSlice := range cScheduler.CollectorTopicMap {
					topicsSliceLen := len(topicsSlice)
					if len(deletedTopicList) != 0 {
						for _, topic := range deletedTopicList {
							var deleteTopic string
							hasTopicIdx := sort.Search(topicsSliceLen, func(i int) bool { return topic <= topicsSlice[i] })
							if hasTopicIdx != topicsSliceLen {
								deleteTopic = topicsSlice[hasTopicIdx]
								for i := 0; i < topicsSliceLen; i++ {
									if topicsSlice[i] == deleteTopic {
										topicsSlice = append(topicsSlice[:i], topicsSlice[i+1:]...)
										i--
									}
								}
							}
						}
						cScheduler.CollectorTopicMap[idx] = topicsSlice
					}
					if len(newTopicList) != 0 {
						if len(topicsSlice) == currentMaxHostCount {
							continue
						} else{
							for _, topic := range newTopicList {
								if len(leftNewTopicList) == 0 {
									break
								}
								topicsSlice = append(topicsSlice, topic)
								leftNewTopicList = leftNewTopicList[1:len(leftNewTopicList)]
								if len(topicsSlice) == currentMaxHostCount {
									cScheduler.CollectorTopicMap[idx] = topicsSlice
									break
								}
							}
						}
						newTopicList = leftNewTopicList
					}
					_ = etcd.GetInstance().WriteMetric(fmt.Sprintf("topicschedule/%d/topics", idx), MergeTopicsToOneString(cScheduler.CollectorTopicMap[idx]), false)
				}
				if len(leftNewTopicList) != 0 {
					newCollectorCount := CalculateNumberOfCollector(len(leftNewTopicList), currentMaxHostCount)
					cScheduler.CollectorCount += newCollectorCount
					newTopicMap := MakeCollectorTopicMap(leftNewTopicList, newCollectorCount, currentMaxHostCount)
					for _, topicMap := range newTopicMap {
						idx := len(cScheduler.CollectorTopicMap) + 1
						cScheduler.CollectorTopicMap[idx] = topicMap
						_ = etcd.GetInstance().WriteMetric(fmt.Sprintf("topicschedule/%d/topics", idx), MergeTopicsToOneString(cScheduler.CollectorTopicMap[idx]), false)
					}
				}
			}
			break
		default:
			createdCollectorNumber := len(cScheduler.cm.CollectorManageSlice)
			switch {
			// TODO Divide logic for scheduling time set config
			case cScheduler.CollectorCount == createdCollectorNumber :
				optimaCollectNumber := CalculateNumberOfCollector(len(currentTopicsState), currentMaxHostCount)
				if cScheduler.CollectorCount > optimaCollectNumber {
					cScheduler.CollectorCount = optimaCollectNumber
				}
			case cScheduler.CollectorCount > createdCollectorNumber :
				cScheduler.CheckScaling = true
				cScheduler.ScalePolicy = "ScaleUp"
				cScheduler.ScaleCount = cScheduler.CollectorCount - createdCollectorNumber
			case cScheduler.CollectorCount < createdCollectorNumber :
				cScheduler.CheckScaling = true
				cScheduler.ScalePolicy = "ScaleDown"
				cScheduler.ScaleCount =  createdCollectorNumber - cScheduler.CollectorCount
			}
		}

		beforeTopicsState = currentTopicsState
		currentTopicsState = GetAllTopicBySort(kafkabroker.GetInstance().GetAllTopics())
		beforeMaxHostCount = currentMaxHostCount
		currentMaxHostCount = config.GetInstance().Monitoring.MaxHostCount
		time.Sleep(2 * time.Second)
	}
}

func (cScheduler CollectorScheduler) Scaler() {
	for{
		if cScheduler.CheckScaling {
			switch cScheduler.ScalePolicy {
			case "ScaleUp":
				for i := 0; i < cScheduler.ScaleCount; i++ {
					cScheduler.cm.CreateCollector()
				}
				break
			case "ScaleDown":
				for i := 0; i < cScheduler.ScaleCount; i++ {
					cScheduler.cm.StopCollector()
				}
				break
			}
			cScheduler.CheckScaling = false
		}
		time.Sleep(time.Duration(config.GetInstance().Monitoring.SchedulingInterval)*time.Second)
	}
}
