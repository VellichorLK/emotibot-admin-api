package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var file string
var output string
var perspective string

type ParseConfig struct {
	Filepath    string
	Perspective []rune
}

func Parse(conf *ParseConfig) (map[string][]byte, error) {
	f, err := os.Open(conf.Filepath)
	if err != nil {
		return nil, fmt.Errorf("open input failed, %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv content failed, %v", err)
	}

	logics := map[string][]string{}
	for i, cols := range rows {
		cols[0] = strings.TrimSpace(strings.Trim(cols[0], "#"))
		cols[1] = strings.TrimSpace(cols[1])
		//skip header
		if i == 0 && cols[0] == "原数据" {
			continue
		}
		if len(cols) < 2 {
			return nil, fmt.Errorf("invalid format, row %d should have at least 2 columns", i+1)
		}
		data := bytes.Runes([]byte(cols[0]))
		tag := bytes.Runes([]byte(cols[1]))
		perspective := data[0]
		if isValidPerspective(perspective) {
			return nil, fmt.Errorf("invalid format, row %d column 1 starting string should be 0 or 1 but not %s", i+1, string(perspective))
		}
		if !containPerspective(perspective, conf.Perspective) {
			continue
		}
		if tag[0] != perspective {
			return nil, fmt.Errorf("invalid format, row %d column 1 & 2 should have the same starting rune", i+1)
		}
		processedText := strings.Replace(string(data[1:]), " ", "", -1)
		processedTag := strings.TrimSpace(string(tag[1:]))

		contents, _ := logics[processedTag]
		contents = append(contents, processedText)
		logics[processedTag] = contents
	}

	ruleBuf := &bytes.Buffer{}
	ruleCsv := csv.NewWriter(ruleBuf)
	ruleCsv.Write([]string{"group_name", "method", "score", "rule", "description", "logic_list"})
	logicBuf := &bytes.Buffer{}
	logicCsv := csv.NewWriter(logicBuf)
	logicCsv.Write([]string{"name", "rule", "distance", "range_int", "range_type"})
	intentBuf := &bytes.Buffer{}
	intentCsv := csv.NewWriter(intentBuf)
	intentCsv.Write([]string{"tag", "asr_text"})
	keywordBuf := &bytes.Buffer{}
	keywordCsv := csv.NewWriter(keywordBuf)
	keywordCsv.Write([]string{"tag", "word"})
	_, fileName := path.Split(conf.Filepath)
	pf := strings.TrimRight(fileName, path.Ext(fileName))

	fmt.Println("logics number: ", len(logics))
	for logicName, texts := range logics {
		ruleCsv.Write([]string{pf, "1", "0", logicName, logicName, logicName})
		logicCsv.Write([]string{logicName, logicName, "0", "", ""})
		for _, text := range texts {
			intentCsv.Write([]string{logicName, text})
		}
	}
	ruleCsv.Flush()
	logicCsv.Flush()
	intentCsv.Flush()
	keywordCsv.Flush()
	quickRead := func(reader io.Reader) []byte {
		data, _ := ioutil.ReadAll(reader)
		return data
	}
	results := map[string][]byte{
		"rule":    quickRead(ruleBuf),
		"logic":   quickRead(logicBuf),
		"intent":  quickRead(intentBuf),
		"keyword": quickRead(keywordBuf),
	}
	return results, nil
}

func Write(filepath string, content []byte) (int, error) {
	_, err := os.Open(filepath)
	if os.IsExist(err) {
		return 0, fmt.Errorf("file already exist in %s", filepath)
	}
	f, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("file can not be created, %v", err)
	}

	return f.Write(content)
}

func isValidPerspective(perspective rune) bool {
	return (perspective != '0' && perspective != '1')
}

func containPerspective(perspective rune, perspectivies []rune) bool {
	for _, p := range perspectivies {
		if perspective == p {
			return true
		}
	}
	return false
}

func NewParseConfig() *ParseConfig {
	parseConfig := &ParseConfig{
		Filepath:    file,
		Perspective: []rune{},
	}
	perspectivies := strings.Split(perspective, ",")
	for _, p := range perspectivies {
		runes := bytes.Runes([]byte(p))
		if len(runes) != 1 {
			log.Fatal("argument p is not valid")
		}
		parseConfig.Perspective = append(parseConfig.Perspective, runes[0])
	}

	return parseConfig
}
func main() {
	flag.StringVar(&file, "f", "", "required -- input csv file path")
	flag.StringVar(&output, "o", "./", "output folder.")
	flag.StringVar(&perspective, "p", "0", "extract role, 0 -- staff 1 -- customer, use comma to indicate both(ex: 0,1)")
	flag.Parse()

	if file == "" {
		log.Fatal("require input directive")
	}
	files, err := Parse(NewParseConfig())
	if err != nil {
		log.Fatal("parse failed, ", err)
	}

	for name, content := range files {
		_, err = Write("./"+name+".csv", content)
		if err != nil {
			log.Fatal("write failed, ", err)
		}
	}
}
