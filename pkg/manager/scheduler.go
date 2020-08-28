package manager

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/kafka"
	"github.com/cloud-barista/cb-dragonfly/pkg/localstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

type CollectorScheduler struct {
	cm *CollectManager
	tm *TopicManager
}

func NewCollectorScheduler(manager *CollectManager) (*CollectorScheduler, error) {

	cScheduler := CollectorScheduler{
		tm:TopicMangerInstance(),
		cm:manager,
	}

	return &cScheduler, nil
}

func (cScheduler CollectorScheduler) Scheduler() {

	currentTopicsState := util.GetAllTopicBySort(kafka.GetInstance().GetAllTopics())
	beforeTopicsState := currentTopicsState
	beforeMaxHostCount, _ := strconv.Atoi(localstore.GetInstance().StoreGet(types.MONCONFIG + "/"+"max_host_count"))
	currentMaxHostCount := beforeMaxHostCount

	topicListChanged := !cmp.Equal(beforeTopicsState, currentTopicsState)
	maxHostCountChanged := !(beforeMaxHostCount == currentMaxHostCount)
	// Init
	cScheduler.tm.SetCollectorPerTopic(currentTopicsState, currentMaxHostCount)
	cScheduler.NeedCollectorScaleInOut()
	cScheduler.SendTopicsToCollectors()

	for {
		aggreTime, _ := strconv.Atoi(localstore.GetInstance().StoreGet(types.MONCONFIG + "/"+"collector_interval"))
		time.Sleep(time.Duration(aggreTime) * time.Second)
		switch {
		case !maxHostCountChanged && !topicListChanged :
			break
		case maxHostCountChanged :
			err := cScheduler.tm.DeleteAllTopicsInfo()
			if err != nil {
				logrus.Debug(err)
			}
			cScheduler.tm.SetCollectorPerTopic(currentTopicsState, currentMaxHostCount)
			cScheduler.NeedCollectorScaleInOut()
			break
		case topicListChanged :
			if !cScheduler.NeedRebalancedTopics(currentTopicsState, currentMaxHostCount) {
				deletedTopicList, newTopicList := cScheduler.ReturnDiffTopics(beforeTopicsState, currentTopicsState)
				err := cScheduler.tm.DeleteTopics(deletedTopicList)
				if err != nil {
					logrus.Debug(err)
				}
				err = cScheduler.tm.AddNewTopics(newTopicList, currentMaxHostCount)
				if err != nil {
					logrus.Debug(err)
				}
			}
			cScheduler.NeedCollectorScaleInOut()
			break
		}
		cScheduler.SendTopicsToCollectors()
		beforeTopicsState = currentTopicsState
		currentTopicsState = util.GetAllTopicBySort(kafka.GetInstance().GetAllTopics())
		fmt.Println(fmt.Sprintf("##### %s : %s #####", "All topics from kafka", currentTopicsState))
		beforeMaxHostCount = currentMaxHostCount
		currentMaxHostCount, _ = strconv.Atoi(localstore.GetInstance().StoreGet(types.MONCONFIG + "/"+"max_host_count"))

		topicListChanged = !cmp.Equal(beforeTopicsState, currentTopicsState)
		maxHostCountChanged = !(beforeMaxHostCount == currentMaxHostCount)
	}
}

func (cScheduler CollectorScheduler) SendTopicsToCollectors() {
	for idx, cAddrList := range cScheduler.cm.CollectorGroupManageMap {
		for _, cAddr := range cAddrList {
			(*cAddr).Ch <- localstore.GetInstance().StoreGet(fmt.Sprintf("%s/%d", types.COLLECTORGROUPTOPIC, idx))
		}
	}
}

func (cScheduler CollectorScheduler) NeedCollectorScaleInOut() {
	var err error
	var idealCollectorCnt int
	if len(cScheduler.tm.IdealCollectorPerAgentCntSlice) == 0 {
		idealCollectorCnt = 1
	} else {
		idealCollectorCnt = len(cScheduler.tm.IdealCollectorPerAgentCntSlice)
	}
	scaleCnt := idealCollectorCnt - len(cScheduler.cm.CollectorGroupManageMap)
	if scaleCnt != 0 {
		for needScalingCnt := scaleCnt; needScalingCnt != 0; {
			if needScalingCnt > 0 {
				err = cScheduler.cm.CreateCollectorGroup()
				needScalingCnt--
			} else {
				err = cScheduler.cm.StopCollectorGroup()
				needScalingCnt++
			}
			if err != nil {
				logrus.Debug(err)
			}
		}
	}
}

func (cScheduler CollectorScheduler) ReturnDiffTopics(beforeTopics []string, currentTopics []string) ([]string, []string) {
	return util.ReturnDiffTopicList(beforeTopics, currentTopics),  util.ReturnDiffTopicList(currentTopics, beforeTopics)
}

func (cScheduler CollectorScheduler) NeedRebalancedTopics(currentTopicsState []string, currentMaxHostCount int) bool {
	if len(cScheduler.tm.IdealCollectorPerAgentCntSlice)  == ((len(currentTopicsState) / currentMaxHostCount) + 1)  {
		return false
	} else {
		err := cScheduler.tm.DeleteAllTopicsInfo()
		if err != nil {
			logrus.Debug(err)
		}
		cScheduler.tm.SetCollectorPerTopic(currentTopicsState, currentMaxHostCount)
		return true
	}
}
