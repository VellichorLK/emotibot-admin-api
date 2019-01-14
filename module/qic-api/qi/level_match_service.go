package qi

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

//error message
var (
	ErrNoArgument      = errors.New("Need arguments")
	ErrTimeoutSet      = errors.New("timeout must be larger than zero")
	ErrWrongExpression = errors.New("wrong conversation flow expression")
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
	Position   int
	Range      int
}

//ExprNode is used to transform expression to node struct
type ExprNode struct {
	withNot bool
	isThen  bool
	uuid    string
}

//ConFlowCriteria is conversation flow matching critera
type ConFlowCriteria struct {
	ID         uint64
	Repeat     int
	Expression string

	startMust bool
	nodes     []*ExprNode
}

//RuleCriteria is criteria for rule level
type RuleCriteria struct {
	ID     uint64
	Min    int
	Score  int
	Method int
	CFIDs  []uint64
}

//RuleMatchedResult is used to return for rule check result
type RuleMatchedResult struct {
	Valid bool
	Score int //plus or minus
}

//RuleGrpCredit is the result of the segments
type RuleGrpCredit struct {
	RuleGrpID                uint64
	Plus                     int
	MatchedRules             []uint64
	MatchedConversationFlows []uint64
	MatchedSentenceGrps      []uint64
	MatchedSentences         []uint64
	MatchedTags              []uint64
}

//FlowExpressionToNode converts the conversation flow expression to node
func (c *ConFlowCriteria) FlowExpressionToNode() error {

	token := strings.Split(c.Expression, " ")
	numOfToken := len(token)
	stack := make(chan string, 99999)

	if numOfToken < 2 {
		return ErrWrongExpression
	}

	lToken := strings.ToLower(token[0])
	switch lToken {
	case "if":
	case "must":
		c.startMust = true
	default:
		return ErrWrongExpression
	}

	stack <- lToken

	for i := 1; i < numOfToken; i++ {

		lToken := strings.ToLower(token[i])
		switch lToken {
		case "not":
			stack <- lToken
		case "and":
			fallthrough
		case "then":
			if len(stack) != 0 {
				return ErrWrongExpression
			}
			stack <- lToken
		default:
			if len(stack) == 0 {
				return ErrWrongExpression
			}
			n := &ExprNode{}
			numOfStack := len(stack)
			for j := 0; j < numOfStack; j++ {
				last := <-stack
				switch last {
				case "and":
				case "not":
					n.withNot = !n.withNot
				case "then":
					n.isThen = true
				case "if":
				case "must":
				default:
					return ErrWrongExpression
				}
			}
			n.uuid = token[i]
			c.nodes = append(c.nodes, n)
		}

	}

	if len(stack) != 0 {
		return ErrWrongExpression
	}
	return nil
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
	c map[uint64]*SenGroupCriteria, segments []*ASRSegment) (map[uint64][]int, error) {
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
							//checkNext = true
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

//ConversationFlowMatch checks one conversation at one time and give whether this convseration flow is meet or not
//parameters:
//matchSgID is the segment index for each sentence group that is matched
//senGrpCriteria is the criteria for each sentence group
//cfCriteria is the conversation flow criteria
//senGrpUUIDMapID is the map from uuid to id in sentence group
//totalSeg is the total lines in this user input context
func ConversationFlowMatch(matchSgID map[uint64][]int, senGrpCriteria map[uint64]*SenGroupCriteria,
	cfCriteria *ConFlowCriteria, senGrpUUIDMapID map[string]uint64, totalSeg int) (matched bool, err error) {

	if cfCriteria == nil {
		return
	}
	//empty the node in case for reuse
	cfCriteria.nodes = []*ExprNode{}
	cfCriteria.startMust = false
	//transform the expression to node struct
	err = cfCriteria.FlowExpressionToNode()
	if err != nil {
		logger.Error.Printf("Transform expresionn %s failed. %s\n", cfCriteria.Expression, err)
		return
	}

	//copy the matched sentence group id for later use
	copyMatchSgID := make(map[uint64][]int, len(matchSgID))
	for k, v := range matchSgID {
		s := make([]int, len(v))
		copy(s, v)
		copyMatchSgID[k] = s
	}

	for z := 0; z < cfCriteria.Repeat; z++ {
		var lastSegmentIdx int
		useFirstSeg := true
		//check for each sentence group which is already in order
		for idx, v := range cfCriteria.nodes {

			id, ok := senGrpUUIDMapID[v.uuid]
			if !ok {
				logger.Error.Printf("Cannot find uuid %s with its id\n", v.uuid)
				return
			}

			criteria, ok := senGrpCriteria[id]
			if !ok {
				logger.Error.Printf("Cannot find sentence group %d with its information\n", id)
				return
			}

			matchThisSenGrp := true
			numOfMatchedSeg := len(copyMatchSgID[id])

			if v.withNot {
				if numOfMatchedSeg != 0 {
					matchThisSenGrp = false
				}
			} else {
				if numOfMatchedSeg == 0 {
					matchThisSenGrp = false
				}
			}

			if idx == 0 {
				//no match at begining
				if !matchThisSenGrp {
					//starts with if
					if !cfCriteria.startMust {
						matched = true
					}
					return
				}

				switch criteria.Position {
				//must start in n words
				case 0:
					if copyMatchSgID[id][0] > criteria.Range {
						return
					}
				//must ends with this sentence group in the n last words
				case 1:
					if totalSeg-copyMatchSgID[id][numOfMatchedSeg-1] > criteria.Range {
						return
					}
					useFirstSeg = false
				//no assigned
				default:
				}
			} else {
				if !matchThisSenGrp {
					return
				}
				if v.withNot {
					continue
				}
				//check then scenario
				if criteria.Range > 0 {
					segIdx := copyMatchSgID[id][0]
					if segIdx-lastSegmentIdx > criteria.Range || segIdx < lastSegmentIdx {
						return
					}
				} else {
					if v.isThen {
						if copyMatchSgID[id][0] < lastSegmentIdx {
							return
						}
					}
				}
			}

			if useFirstSeg {
				lastSegmentIdx = copyMatchSgID[id][0]
				copyMatchSgID[id] = copyMatchSgID[id][1:]
			} else {
				lastSegmentIdx = copyMatchSgID[id][numOfMatchedSeg-1]
				copyMatchSgID[id] = copyMatchSgID[id][:numOfMatchedSeg]
			}
		}
	}
	matched = true
	return
}

//RuleMatch used to check whether the rule level meets. gives the map that the rule id meets the criterion and its plus score
//paramters:
//cfMatchID is the map recording the conversation flow id which meets the criterion
func RuleMatch(cfMatchID map[uint64]bool, criteria map[uint64]*RuleCriteria) (map[uint64]*RuleMatchedResult, int, error) {
	resp := make(map[uint64]*RuleMatchedResult, len(criteria))
	var totalScore int
	for ruleID, criterion := range criteria {
		var count int
		var matched bool
		var plus int
		for _, cfID := range criterion.CFIDs {
			if v, ok := cfMatchID[cfID]; ok && v {
				count++
			}
		}
		if count >= criterion.Min {
			matched = true
		}
		if criterion.Method == int(methodStringToCode["negative"]) {
			matched = !matched
		}

		if matched {
			if criterion.Score > 0 {
				plus = plus + criterion.Score
			}
		} else {
			if criterion.Score < 0 {
				plus = plus + criterion.Score
			}
		}
		totalScore += plus
		resp[ruleID] = &RuleMatchedResult{Valid: matched, Score: plus}
	}
	return resp, totalScore, nil
}

//RuleGroupCriteria gives the result of the criteria used to the group
//the parameter timeout is used to wait for cu module result
func RuleGroupCriteria(ruleGroup uint64, segments []*ASRSegment, timeout time.Duration) ([]string, error) {
	numOfLines := len(segments)
	if numOfLines == 0 {
		return nil, ErrNoArgument
	}

	//get the relation table from RuleGroup to Tag
	levels, _, err := GetLevelsRel(LevRuleGroup, LevTag, []uint64{ruleGroup})
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

	numOfSens := len(levels[LevSentence])

	//extract the sentence id and tag id
	senIDs := make([]uint64, 0, numOfSens)
	tagIDs := make([]uint64, 0, numOfSens)
	for sID, tIDs := range levels[LevSentence] {
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
	senMatchDat, err := SentencesMatch(segMatchedTag, levels[LevSentence])
	if err != nil {
		logger.Warn.Printf("doing sentence  match failed.%s\n", err)
		return nil, err
	}

	//extract the sentence group id
	conContainSenGrp := levels[LevConversation]
	senGrpIDs := make([]uint64, 0)
	cfIDs := make([]uint64, 0, len(conContainSenGrp))
	for cfID, senGrpIDList := range conContainSenGrp {
		senGrpIDs = append(senGrpIDs, senGrpIDList...)
		cfIDs = append(cfIDs, cfID)
	}

	//get the sentence group information for condition usage
	sgFilter := &model.SentenceGroupFilter{ID: senGrpIDs, IsDelete: -1, Position: -1, Role: -1}
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
	senGrpCriteria := make(map[uint64]*SenGroupCriteria)
	senGrpUUIDMapID := make(map[string]uint64, numOfSenGrp)

	for i := 0; i < numOfSenGrp; i++ {
		id := uint64(senGrp[i].ID)
		var criterion SenGroupCriteria
		if senIDs, ok := senGrpContainSen[id]; ok {
			senGrpCriteria[id] = &criterion
			senGrpCriteria[id].ID = id
			senGrpCriteria[id].Role = senGrp[i].Role
			senGrpCriteria[id].Range = senGrp[i].Distance
			senGrpCriteria[id].Position = senGrp[i].Position
			senGrpCriteria[id].SentenceID = senIDs
			senGrpUUIDMapID[senGrp[i].UUID] = id
		} else {
			logger.Warn.Printf("No sentence group id %d in sentence group table, but exist in relation table\n", id)
		}
	}

	//do the check, sentence group
	matchSgID, err := SentenceGroupMatch(senMatchDat, senGrpCriteria, segments)
	if err != nil {
		logger.Warn.Printf("doing sentence group match failed.%s\n", err)
		return nil, err
	}

	//get the conversation flow inforamtion
	cfFilter := &model.ConversationFlowFilter{ID: cfIDs, IsDelete: -1}
	_, cfInfo, err := GetConversationFlowsBy(cfFilter)
	if err != nil {
		logger.Error.Printf("get conversation flow failed.%s\n", err)
		return nil, err
	}

	//sorting the matched segment index
	for _, segIdxs := range matchSgID {
		sort.Ints(segIdxs)
	}

	matchCFID := make(map[uint64]bool)
	//doing check for each conversation flow
	for i := 0; i < len(cfInfo); i++ {
		var c ConFlowCriteria
		c.ID = uint64(cfInfo[i].ID)
		c.Expression = cfInfo[i].Expression

		cfMatched, err := ConversationFlowMatch(matchSgID, senGrpCriteria, &c, senGrpUUIDMapID, numOfLines)
		if err != nil {
			logger.Error.Printf("getting the conversation flow match failed. %s\n", err)
			return nil, err
		}
		if cfMatched {
			matchCFID[c.ID] = true
		}
	}

	ruleGrpContainRule := levels[LevRuleGroup]
	ruleGrpIDs := make([]uint64, 0, len(ruleGrpContainRule))
	ruleIDs := make([]uint64, 0, len(ruleGrpContainRule))
	for rGrpID, ruleList := range ruleGrpContainRule {
		ruleGrpIDs = append(ruleGrpIDs, rGrpID)
		ruleIDs = append(ruleIDs, ruleList...)
	}

	ruleFilter := &model.ConversationRuleFilter{ID: ruleIDs, IsDeleted: -1, Severity: -1}
	_, rules, err := GetConversationRulesBy(ruleFilter)
	if err != nil {
		logger.Error.Printf("get the rules failed.%s\n", err)
		return nil, err
	}
	ruleCriteria := make(map[uint64]*RuleCriteria)
	for _, v := range rules {
		c := &RuleCriteria{}
		c.ID = uint64(v.ID)
		c.Method = int(v.Method)
		c.Min = v.Min
		c.Score = v.Score
		for _, cfID := range v.Flows {
			c.CFIDs = append(c.CFIDs, uint64(cfID.ID))
		}
		ruleCriteria[c.ID] = c
	}

	matchRule, totalScore, err := RuleMatch(matchCFID, ruleCriteria)
	if err != nil {
		logger.Error.Printf("rule level match failed.%s\n", err)
		return nil, err
	}

	_ = matchRule
	_ = totalScore

	return nil, nil
}
