package qi

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

//error message
var (
	ErrNoArgument = errors.New("Need arguments")
	ErrTimeoutSet = errors.New("timeout must be larger than zero")
)

//MatchedData stores the index of input and the matched ID (tag id) and its relative data
type MatchedData struct {
	Index   uint64
	Matched map[uint64]*logicaccess.AttrResult
	lock    sync.Mutex
}

//ContainRelation is relation between parent and child
type ContainRelation struct {
	parent uint64
	child  []uint64
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
	sentences []string, wg *sync.WaitGroup, collected []*MatchedData) {
	defer wg.Done()
	numOfData := len(collected) + 1
	for {
		select {
		case id, more := <-target:
			if !more {
				return
			}
			pr, err := BatchPredict(id, Threshold, sentences)
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

//TagMatch checks each sentence for each tags
//return value: slice of matchData gives the each sentences and its matched tag and matched data
func TagMatch(tags []uint64, sentences []string, timeout time.Duration) ([]*MatchedData, error) {

	numOfTags := len(tags)
	numOfCtx := len(sentences)

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
		matches[i].Index = uint64(i + 1)
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
		go worker(ctx, target, sentences, &wg, matches)
	}

	wg.Wait()

	return matches, ctx.Err()
}

//SentencesMatch gives the sentence id that is matched by which segment
//c is the argument map[sentence id][]tag id.
func SentencesMatch(m []*MatchedData, c map[uint64][]uint64) (map[uint64][]uint64, error) {

	resp := make(map[uint64][]uint64, len(c))

	//for loop the criteria for each tag in each sentence
	for sID, tagIDs := range c {
		//compare the given matched tags in each segement to the criteria
		for _, d := range m {
			if d != nil && len(d.Matched) > 0 {
				numOfChild := len(tagIDs)
				var count int
				//check whether this segment match all tags
				for _, tagID := range tagIDs {
					if _, ok := d.Matched[tagID]; !ok {
						break
					}
					count++
				}
				if count == numOfChild {
					resp[sID] = append(resp[sID], d.Index)
				}
			}
		}
	}
	return resp, nil
}

//RuleGroupCriteria gives the result of the criteria used to the group
//the parameter timeout is used to wait for cu module result
func RuleGroupCriteria(ruleGroup uint64, sentences []string, timeout time.Duration) ([]string, error) {
	numOfLines := len(sentences)
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

	//do the checking, tag match
	tagMatchDat, err := TagMatch(tagIDs, sentences, timeout)
	if err != nil {
		return nil, err
	}
	if len(tagMatchDat) != numOfLines {
		logger.Error.Printf("get less tag match sentence. %d\n", numOfLines)
		return nil, err
	}

	//do the checking, sentence match
	senMatchDat, err := SentencesMatch(tagMatchDat, levels[sentenceLev])
	if err != nil {
		return nil, err
	}
	_ = senMatchDat
	return nil, nil
}
