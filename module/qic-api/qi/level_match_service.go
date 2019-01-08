package qi

import (
	"context"
	"errors"
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
	Index   int
	Matched map[uint64]*logicaccess.AttrResult
	lock    sync.Mutex
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

func collector(ctx context.Context, response <-chan *logicaccess.AttrResult, collected []MatchedData) {

	numOfData := len(collected) + 1
	for {
		select {
		case d, more := <-response:
			if !more {
				return
			}
			if d != nil && d.SentenceID > 0 && d.SentenceID < numOfData {
				idx := d.SentenceID - 1
				collected[idx].Matched[d.Tag] = d
			}
		case <-ctx.Done():
			return
		}
	}
}

//TagMatch checks each sentence for each tags
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
		matches[i].Index = i + 1
	}

	//start to input the target tag id
	for _, v := range tags {
		target <- v
	}
	close(target)

	//call goroutine to do job concurrency
	for i := 0; i < Concurrency; i++ {
		go worker(ctx, target, sentences, &wg, matches)
	}

	wg.Wait()

	return matches, ctx.Err()
}
