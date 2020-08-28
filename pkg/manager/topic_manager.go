package manager

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/localstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

type TopicManager struct {
	IdealCollectorGroupMap map[int] []string
	IdealCollectorPerAgentCntSlice [] int
}

var once sync.Once
var topicManager TopicManager

func TopicMangerInit() {
	topicManager.IdealCollectorGroupMap = map[int][]string{}
	topicManager.IdealCollectorPerAgentCntSlice = []int{}
}

func TopicMangerInstance() *TopicManager {
	once.Do(func() {
		TopicMangerInit()
	})
	return &topicManager
}

func (t *TopicManager) SetCollectorPerTopic(topicList [] string, maxHostCount int) {
	t.IdealCollectorGroupMap, t.IdealCollectorPerAgentCntSlice = util.MakeCollectorTopicMap(topicList, maxHostCount)
	if len(t.IdealCollectorGroupMap) == 0 && len(t.IdealCollectorPerAgentCntSlice) == 0 {
		_ = localstore.GetInstance().StorePut(fmt.Sprintf("%s/%d", types.COLLECTORGROUPTOPIC, 0), "")
		return
	}
	for collectorIdx, collectorTopics := range t.IdealCollectorGroupMap {
		for i:=0; i< len(collectorTopics); i++ {
			_ = localstore.GetInstance().StorePut(fmt.Sprintf("%s/%s", types.TOPIC, collectorTopics[i]), strconv.Itoa(collectorIdx))
		}
		_ = localstore.GetInstance().StorePut(fmt.Sprintf("%s/%d", types.COLLECTORGROUPTOPIC, collectorIdx), util.MergeTopicsToOneString(collectorTopics))
	}
}

func (t *TopicManager) DeleteAllTopicsInfo() error {
	err := localstore.GetInstance().StoreDelList(fmt.Sprintf("%s/", types.COLLECTORGROUPTOPIC))
	if err != nil {
		return err
	}
	return nil
}

func (t *TopicManager) DeleteTopics(deletedTopicList []string) error {
	if len(deletedTopicList) == 0 {
		return nil
	}
	changedCollectorMapIdx := map[string] []string{}
	for _, topic := range deletedTopicList {
		collectorIdx := localstore.GetInstance().StoreGet(fmt.Sprintf("%s/%s", types.TOPIC, topic))
		topicStrings := localstore.GetInstance().StoreGet(fmt.Sprintf("%s/%s", types.COLLECTORGROUPTOPIC, collectorIdx))
		if _, ok := changedCollectorMapIdx[collectorIdx]; !ok {
			changedCollectorMapIdx[collectorIdx] = util.SplitOneStringToTopicsSlice(topicStrings)
		}
		changedCollectorMapIdx[collectorIdx] = util.ReturnDiffTopicList(changedCollectorMapIdx[collectorIdx], [] string{topic})
		err := localstore.GetInstance().StoreDelete(fmt.Sprintf("%s/%s", types.TOPIC, topic))
		if err != nil {
			return err
		}
		idx, _ := strconv.Atoi(collectorIdx)
		t.IdealCollectorPerAgentCntSlice[idx] -= 1
	}
	for key, _ := range changedCollectorMapIdx {
		err := localstore.GetInstance().StorePut(fmt.Sprintf("%s/%s", types.COLLECTORGROUPTOPIC, key), util.MergeTopicsToOneString(changedCollectorMapIdx[key]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TopicManager) AddNewTopics(newTopicList []string, maxHostCount int) error {
	if len(newTopicList) == 0 {
		return nil
	}
	for _, topic := range newTopicList {
		for collectorIdx, collectorTopicCnt := range t.IdealCollectorPerAgentCntSlice {
			if collectorTopicCnt < maxHostCount {
				err := localstore.GetInstance().StorePut(fmt.Sprintf("%s/%s", types.TOPIC, topic), strconv.Itoa(collectorIdx))
				if err != nil {
					return err
				}
				err = localstore.GetInstance().StorePut(fmt.Sprintf("%s/%d", types.COLLECTORGROUPTOPIC, collectorIdx),localstore.GetInstance().StoreGet(fmt.Sprintf("%s/%d", types.COLLECTORGROUPTOPIC, collectorIdx))+"&"+topic)
				if err != nil {
					return err
				}
				t.IdealCollectorPerAgentCntSlice[collectorIdx] += 1
				break
			}
		}
	}
	return nil
}
