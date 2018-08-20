package data

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"emotibot.com/emotigo/pkg/logger"
	"github.com/gonum/blas/blas64"
)

var StopWordList map[string]int

func SetStopWords(path string) {
	StopWordList = make(map[string]int)
	file, err := os.Open(path + "resources/stoplist.txt")
	if err != nil {
		logger.Error.Println("Cannot load from file path stoplist.txt")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		StopWordList[scanner.Text()] = 1
	}
}

type Vector []float64

var (
	StopWord    map[string]struct{}
	Lv0StopWord map[string]struct{}
	Lv1StopWord map[string]struct{}
	TimeWord    map[string]struct{}
	Punctuation map[string]struct{}
	IDFCache    sync.Map
	Word2Vec    sync.Map
	dim         uint
	size        uint
	initialize  bool
)

func InitializeWord2Vec(path string) bool {
	initialize = true
	loadModel(path)
	loadStopWords(path)
	loadPunctuation(path)
	loadIDFCache(path)
	return initialize
}

func GetSentenceVector(keywords, token []string) Vector {
	kwSentenceVector := getEmbedding(keywords)
	tokenSentenceVector := getEmbedding(token)
	embedding := make(Vector, dim)

	var i uint
	for i = 0; i < dim; i++ {
		if kwSentenceVector != nil && tokenSentenceVector != nil {
			embedding[i] = float64(0.8*tokenSentenceVector[i] + 0.2*kwSentenceVector[i])
		} else if kwSentenceVector == nil && tokenSentenceVector != nil {
			embedding[i] = float64(tokenSentenceVector[i])
		} else {
			return nil
		}
	}
	return embedding
}

func getEmbedding(words []string) Vector {
	if words == nil || len(words) == 0 {
		return nil
	}
	var (
		cnt                   uint
		num                   float64
		numRemoved            float64
		sentenceVector        Vector
		sentenceVectorRemoved Vector
	)

	sentenceVector = make(Vector, dim)
	sentenceVectorRemoved = make(Vector, dim)

	for _, word := range words {
		if len(strings.Trim(word, " ")) == 0 {
			continue
		}
		var (
			vector Vector
			weight float64
			remove bool
		)
		vector = getWordVector(word)
		if _, ok := Punctuation[word]; ok {
			weight = 0.01
			remove = true
		} else if _, ok := Lv0StopWord[word]; ok {
			weight = 1.0
		} else if _, ok := Lv1StopWord[word]; ok {
			weight = 1.5
		} else if _, ok := TimeWord[word]; ok {
			weight = 2.0
		} else if _, ok := StopWord[word]; ok {
			weight = 1.5
			remove = true
		} else if matched, _ := regexp.MatchString(".*\\d+.*", word); matched {
			weight = 2.0
		} else {
			tmp, _ := IDFCache.LoadOrStore(word, 10.0)
			weight = tmp.(float64)
		}

		if weight == 0.0 {
			weight = 1.0
		}
		weight = math.Min(10.0, weight)

		if remove {
			numRemoved += weight
			for cnt = 0; cnt < dim; cnt++ {
				sentenceVectorRemoved[cnt] += vector[cnt] * weight
			}
		} else {
			num += weight
			for cnt = 0; cnt < dim; cnt++ {
				sentenceVector[cnt] += vector[cnt] * weight
			}
		}
	}

	if num == 0.0 {
		//logger.Info("[%v] Num == 0. Use the embedding of the words that are supposed to remove in the end.",
		//	words)
		for cnt = 0; cnt < dim; cnt++ {
			sentenceVector[cnt] = sentenceVectorRemoved[cnt] / numRemoved
		}
	} else {
		for cnt = 0; cnt < dim; cnt++ {
			sentenceVector[cnt] /= num
		}
	}
	return sentenceVector
}

func getWordVector(word string) Vector {
	result := getWordVectorItself(word)
	if result == nil {
		result = nilWordVectorHandler(word)
	}
	return result
}

func getWordVectorItself(word string) Vector {
	data, ok := Word2Vec.Load(word)
	if !ok {
		return nil
	}
	return data.(Vector)
}

func nilWordVectorHandler(word string) Vector {
	vector := randomWordVector(dim)
	Word2Vec.LoadOrStore(word, vector)
	return vector
}

func randomWordVector(d uint) Vector {
	var i uint
	var vector Vector
	for i = 0; i < d; i++ {
		vector = append(vector, rand.Float64())
	}
	vector.Normalize()
	return vector
}

func loadStopWords(path string) {
	StopWord = make(map[string]struct{})
	Lv0StopWord = make(map[string]struct{})
	Lv1StopWord = make(map[string]struct{})
	TimeWord = make(map[string]struct{})

	file, err := os.Open(path + "resources/stoplist.txt")
	if err != nil {
		logger.Error.Printf("Read stop list error: %s\n", err)
		initialize = false
		return
	}
	defer file.Close()

	file0, err := os.Open(path + "resources/stoplist_0.txt")
	if err != nil {
		logger.Error.Printf("Read stop list 0 error: %s\n", err)
		initialize = false
		return
	}
	defer file0.Close()

	file1, err := os.Open(path + "resources/stoplist_1.txt")
	if err != nil {
		logger.Error.Printf("Read stop list 1 error:%s\n ", err)
		initialize = false
		return
	}
	defer file1.Close()

	time, err := os.Open(path + "resources/timelist.txt")
	if err != nil {
		logger.Error.Printf("Read time list error: %s\n", err)
		initialize = false
		return
	}
	defer time.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		StopWord[scanner.Text()] = struct{}{}
	}

	scanner0 := bufio.NewScanner(file0)
	for scanner0.Scan() {
		Lv0StopWord[scanner0.Text()] = struct{}{}
	}

	scanner1 := bufio.NewScanner(file1)
	for scanner1.Scan() {
		Lv1StopWord[scanner1.Text()] = struct{}{}
	}

	scannerT := bufio.NewScanner(time)
	for scannerT.Scan() {
		TimeWord[scannerT.Text()] = struct{}{}
	}
}

func loadPunctuation(path string) {
	Punctuation = make(map[string]struct{})

	file, err := os.Open(path + "resources/symbol.txt")
	if err != nil {
		logger.Error.Printf("Read punctuation list error:%s\n ", err)
		initialize = false
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		Punctuation[scanner.Text()] = struct{}{}
	}
}

func loadIDFCache(path string) {
	file, err := os.Open(path + "resources/idf_cache.txt")
	if err != nil {
		logger.Error.Printf("Read idf-cache error: %s\n", err)
		initialize = false
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		num, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			logger.Warn.Printf("Convert to float64 fail: %s\n", line[1])
			continue
		}
		IDFCache.Store(line[0], num)
	}
}

func loadModel(path string) {
	file, err := os.Open(path + "resources/w2v_model.bin")
	if err != nil {
		logger.Error.Printf("Read w2v-model error:%s\n ", err)
		initialize = false
		return
	}
	defer file.Close()

	br := bufio.NewReader(file)
	n, err := fmt.Fscanln(file, &size, &dim)
	if err != nil || n != 2 {
		logger.Error.Printf("Read model size & dimension parameters error:%s\n ", err)
		initialize = false
		return
	}

	raw := make([]float32, size*dim)

	var (
		i    uint
		word string
	)
	for i = 0; i < size; i++ {
		bytes, err := br.ReadBytes(' ')
		if err != nil {
			logger.Error.Printf("Read model data error: %s\n", err)
			initialize = false
			return
		}
		word = string(bytes[:len(bytes)-1])

		v := []float32(raw[dim*i : dim*(i+1)])
		vv := make(Vector, dim)
		if err := binary.Read(br, binary.LittleEndian, v); err != nil {
			logger.Error.Printf("Read word vector error: %s\n", err)
			initialize = false
			return
		}

		for j := range v {
			vv[j] = float64(v[j])
		}

		vv.Normalize()
		Word2Vec.Store(word, vv)

		b, err := br.ReadByte()
		if err != nil {
			if i == size-1 && err == io.EOF {
				break
			}
			logger.Error.Printf("Read w2v-model error:%s\n ", err)
			initialize = false
			return
		}
		if b != byte('\n') {
			if err := br.UnreadByte(); err != nil {
				logger.Error.Printf("Read w2v-model error:%s\n ", err)
				initialize = false
				return
			}
		}
	}
}

func (v Vector) Normalize() {
	w := blas64.Vector{Inc: 1, Data: v}
	total := math.Sqrt(blas64.Dot(len(v), w, w))
	for i := range v {
		v[i] /= total
	}
}
