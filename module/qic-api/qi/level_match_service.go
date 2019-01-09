package qi

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

//error message
var (
	ErrNoArgument = errors.New("Need arguments")
	ErrTimeoutSet = errors.New("timeout must be larger than zero")
)

//MatchedData stores the index of input and the matched ID (tag id) and its relative data
type MatchedData struct {
	Index   int
	Matched map[uint64]*logicaccess.AttrResult
	lock    sync.Mutex
}

//ASRSegment is asr sentence structure
type ASRSegment struct {
	//ID       uint64
	Sentence string
	Speaker  int
}

//SenGroupCriteria is SentenceGroup matching criteria
type SenGroupCriteria struct {
	ID         uint64
	SentenceID []uint64
	Role       int
}

//ConFlowCriteria is conversation flow matching critera
type ConFlowCriteria struct {
	ID         uint64
	expression string
}

//SetData sets the data for thread-safe
func (m *MatchedData) SetData(d *logicaccess.AttrResult) {

	if d != nil && d.SentenceID > 0 {
		m.lock.Lock()
		m.Matched[d.Tag] = d
		m.lock.Unlock()
	}

}

//MatchedIdx stores which index of sentence matchs the target id
type MatchedIdx struct {
	Index     []uint64
	MatchedID uint64
}

//Concurrency sets the number of goroutine used to call cu module
const (
	Concurrency = 5
	Threshold   = 60
)

func worker(ctx context.Context, target <-chan uint64,
	segments []string, wg *sync.WaitGroup, collected []*MatchedData) {
	defer wg.Done()
	numOfData := len(collected) + 1
	for {
		select {
		case id, more := <-target:
			if !more {
				return
			}
			pr, err := BatchPredict(id, Threshold, segments)
			if err != nil {
				logger.Error.Printf("batch predict failed. %s\n", err)
				return
			}

			for i := 0; i < len(pr.Dialogue); i++ {
				v := pr.Dialogue[i]
				if v.SentenceID > 0 && v.SentenceID < numOfData {
					idx := v.SentenceID - 1
					collected[idx].SetData(&v)
				}
			}

			for i := 0; i < len(pr.Keyword); i++ {
				v := pr.Dialogue[i]
				if v.SentenceID > 0 && v.SentenceID < numOfData {
					idx := v.SentenceID - 1
					collected[idx].SetData(&v)
				}
			}

			for i := 0; i < len(pr.UsrResponse); i++ {
				v := pr.Dialogue[i]
				if v.SentenceID > 0 && v.SentenceID < numOfData {
					idx := v.SentenceID - 1
					collected[idx].SetData(&v)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}

//TagMatch checks each segment for each tags
//return value: slice of matchData gives the each sentences and its matched tag and matched data
func TagMatch(tags []uint64, segments []string, timeout time.Duration) ([]*MatchedData, error) {

	numOfTags := len(tags)
	numOfCtx := len(segments)

	if numOfTags == 0 || numOfCtx == 0 {
		return nil, ErrNoArgument
	}
	if timeout <= 0 {
		return nil, ErrTimeoutSet
	}

	//context and channel init
	var wg sync.WaitGroup
	wg.Add(Concurrency)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	target := make(chan uint64, numOfTags)
	defer cancel()

	//init the response structure
	matches := make([]*MatchedData, numOfCtx, numOfCtx)
	for i := 0; i < numOfCtx; i++ {
		matches[i] = &MatchedData{}
		matches[i].Matched = make(map[uint64]*logicaccess.AttrResult)
		matches[i].Index = i + 1
	}

	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
	var lastTag uint64
	//start to input the target tag id
	for _, v := range tags {
		//avoid the duplicate tag
		if lastTag != v {
			target <- v
			lastTag = v
		}
	}
	close(target)

	//call goroutine to do job concurrency
	for i := 0; i < Concurrency; i++ {
		go worker(ctx, target, segments, &wg, matches)
	}

	wg.Wait()

	return matches, ctx.Err()
}

//SentencesMatch gives the sentence id that is matched by which segment index
//c is the argument map[sentence id][]tag id.
//senMatched is matched tag id for each segment
func SentencesMatch(senMatched []map[uint64]bool, c map[uint64][]uint64) (map[uint64][]int, error) {
	//func SentencesMatch(m []*MatchedData, c map[uint64][]uint64) (map[uint64][]int, error) {

	resp := make(map[uint64][]int, len(c))

	//for loop the criteria for each tag in each sentence
	for sID, tagIDs := range c {
		//compare the given matched tags in each segement to the criteria
		for idx, d := range senMatched {
			if len(d) > 0 {
				numOfChild := len(tagIDs)
				var count int
				//check whether this segment match all tags
				for _, tagID := range tagIDs {
					if _, ok := d[tagID]; !ok {
						break
					}
					count++
				}
				if count == numOfChild {
					resp[sID] = append(resp[sID], idx+1)
				}
			}
		}
	}
	return resp, nil
}

//SentenceGroupMatch matches the given matched sentence to sentence group
//matchedSen is matched sentence id and the matched segment id
//c is the sentenceGroup criteria used to judge whether the sentence group is meet
//semgnets is the segments data
//return value: the sentence group id which is meet by which segment index
func SentenceGroupMatch(matchedSen map[uint64][]int,
	c []SenGroupCriteria, segments []*ASRSegment) (map[uint64][]int, error) {
	resp := make(map[uint64][]int, len(c))

	numOfSegs := len(segments)
	//for loop for each sentence group critera
	for _, criteria := range c {

		var checkNext bool
		//for loop for each sentence in sentence group
		for _, sID := range criteria.SentenceID {
			if checkNext {
				break
			}
			//check whether one of the segment is meet the sentence
			if segIdxs, ok := matchedSen[sID]; ok {

				//record which segment meet the sentence
				for _, segIdx := range segIdxs {
					if segIdx-1 < numOfSegs && segments[segIdx-1] != nil {

						s := segments[segIdx-1]

						//FIX ME
						//check the role
						if s.Speaker == criteria.Role {
							resp[criteria.ID] = append(resp[criteria.ID], segIdxs...)
							checkNext = true
							break
						}
					}

				}
			}
		}
	}

	return resp, nil
}

func extractTagMatchedData(tagMatchDat []*MatchedData) []map[uint64]bool {

	numOfData := len(tagMatchDat)
	resp := make([]map[uint64]bool, numOfData, numOfData)
	for i := 0; i < numOfData; i++ {
		resp[i] = make(map[uint64]bool)
		d := tagMatchDat[i]
		if d != nil {
			for tagID := range d.Matched {
				resp[i][tagID] = true
			}
		}
	}

	return resp

}

//ConversationFlowMatch gives the id of conversationFlow that meet the critera
func ConversationFlowMatch(matchSgID map[uint64][]int,
	criteria []*ConFlowCriteria, senGrpUUIDMapID map[string]uint64) (map[uint64]bool, error) {
	for _, c := range criteria {
		if c != nil {
			//c.expression
		}
	}

	return nil, nil
}

//RuleGroupCriteria gives the result of the criteria used to the group
//the parameter timeout is used to wait for cu module result
func RuleGroupCriteria(ruleGroup uint64, segments []*ASRSegment, timeout time.Duration) ([]string, error) {
	numOfLines := len(segments)
	if numOfLines == 0 {
		return nil, ErrNoArgument
	}

	//get the relation table from RuleGroup to Tag
	levels, err := GetLevelsRel(LevRuleGroup, LevTag, []uint64{ruleGroup})
	if err != nil {
		logger.Error.Printf("get level relations failed. %s\n", err)
		return nil, err
	}

	//check the return level
	tagLev := int(LevTag)
	if len(levels) != tagLev {
		logger.Error.Printf("get less relation table. %d\n", tagLev)
		return nil, err
	}

	sentenceLev := int(LevSentence)
	numOfSens := len(levels[sentenceLev])

	//extract the sentence id and tag id
	senIDs := make([]uint64, 0, numOfSens)
	tagIDs := make([]uint64, 0, numOfSens)
	for sID, tIDs := range levels[sentenceLev] {
		senIDs = append(senIDs, sID)
		tagIDs = append(tagIDs, tIDs...)
	}

	//extract the words
	lines := make([]string, 0, numOfLines)
	for _, v := range segments {
		if v != nil {
			lines = append(lines, v.Sentence)
		}
	}

	//do the checking, tag match
	tagMatchDat, err := TagMatch(tagIDs, lines, timeout)
	if err != nil {
		return nil, err
	}
	if len(tagMatchDat) != numOfLines {
		logger.Error.Printf("get less tag match sentence. %d\n", numOfLines)
		return nil, err
	}

	segMatchedTag := extractTagMatchedData(tagMatchDat)
	//do the checking, sentence match
	senMatchDat, err := SentencesMatch(segMatchedTag, levels[sentenceLev])
	if err != nil {
		logger.Warn.Printf("doing sentence  match failed.%s\n", err)
		return nil, err
	}

	//extract the sentence group id
	ConContainSenGrp := levels[LevConversation]
	senGrpIDs := make([]uint64, 0)
	cfIDs := make([]uint64, 0, len(ConContainSenGrp))
	for cfID, senGrpIDList := range ConContainSenGrp {
		senGrpIDs = append(senGrpIDs, senGrpIDList...)
		cfIDs = append(cfIDs, cfID)
	}

	//get the sentence group information for condition usage
	sgFilter := &model.SentenceGroupFilter{ID: senGrpIDs}
	_, senGrp, err := GetSentenceGroupsBy(sgFilter)

	if err != nil {
		logger.Error.Printf("get sentence group info failed.%s\n", err)
		return nil, err
	}
	numOfSenGrp := len(senGrp)
	if numOfSenGrp == 0 {
		logger.Warn.Printf("No sentence group for %d rule group\n", ruleGroup)
		return nil, nil
	}

	//transform the sentence group information to the sentence group critera struct
	senGrpContainSen := levels[LevSenGroup]
	senGrpCriteria := make([]SenGroupCriteria, 0, numOfSenGrp)
	senGrpUUIDMapID := make(map[string]uint64)
	for i := 0; i < numOfSenGrp; i++ {
		var c SenGroupCriteria
		c.ID = uint64(senGrp[i].ID)
		if segIDs, ok := senGrpContainSen[c.ID]; ok {
			c.Role = senGrp[i].Role
			c.SentenceID = segIDs
			senGrpCriteria = append(senGrpCriteria, c)
			senGrpUUIDMapID[senGrp[i].UUID] = c.ID
		} else {
			logger.Warn.Printf("No sentence group id %d in sentence group table, but exist in relation table\n", c.ID)
		}
	}

	//do the check, sentence group
	matchSgID, err := SentenceGroupMatch(senMatchDat, senGrpCriteria, segments)
	if err != nil {
		logger.Warn.Printf("doing sentence group match failed.%s\n", err)
		return nil, err
	}
	_ = matchSgID

	cfFilter := &model.ConversationFlowFilter{ID: cfIDs}
	_, cfInfo, err := GetConversationFlowsBy(cfFilter)
	if err != nil {
		logger.Error.Printf("get conversation flow failed.%s\n", err)
		return nil, err
	}
	numOfCFInfo := len(cfInfo)
	cfCriteria := make([]*ConFlowCriteria, 0, numOfCFInfo)
	for _, info := range cfInfo {
		var c ConFlowCriteria
		c.ID = uint64(info.ID)
		c.expression = info.Expression
		cfCriteria = append(cfCriteria, &c)
	}

	cfMatched, err := ConversationFlowMatch(matchSgID, cfCriteria, senGrpUUIDMapID)
	if err != nil {
		logger.Warn.Printf("doing conversation flow match failed.%s\n", err)
		return nil, err
	}

	_ = cfMatched

	return nil, nil
}
