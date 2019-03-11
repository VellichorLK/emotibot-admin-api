package zhconverter

import (
	"bufio"
	"bytes"
	"os"
    "strings"

	"emotibot.com/emotigo/pkg/logger"
)

const t2sDictFile = "./InitFiles/zh_converter/zh-tw2zh-cn.properties"
const s2tDictFile = "./InitFiles/zh_converter/zh-cn2zh-tw.properties"

var t2sMapping = make(map[string]string)
var s2tMapping = make(map[string]string)

func init() {
	logger.Error.Println("Start loading zh-converter dictionaries")

	t2sDict, err := os.Open(t2sDictFile)
	if err != nil {
		logger.Error.Printf("Load %s failed: %s\n", t2sDictFile, err.Error())
		return
	}
	defer t2sDict.Close()

	s2tDict, err := os.Open(s2tDictFile)
	if err != nil {
		logger.Error.Printf("Load %s failed: %s\n", s2tDictFile, err.Error())
		return
	}
	defer s2tDict.Close()

	var scanner *bufio.Scanner

	// Read zh-tw to zh-cn dictionary
	scanner = bufio.NewScanner(t2sDict)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		dict := strings.Split(text, "=")
		if len(dict) != 2 {
			logger.Warn.Printf("%s: Invalidate dictionary format\n", text)
			continue
		}

		tw := dict[0]
		cn := dict[1]
		t2sMapping[string(tw)] = string(cn)
	}

	// Read zh-tw to zh-cn dictionary
	scanner = bufio.NewScanner(s2tDict)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		dict := strings.Split(text, "=")
		if len(dict) != 2 {
			logger.Warn.Printf("%s: Invalidate dictionary format\n", text)
			continue
		}

		cn := dict[0]
		tw := dict[1]
		s2tMapping[cn] = tw
	}

	logger.Error.Println("zh-converter dictionaries load completed")
}

// Convert Traditional Chinese to Simplified Chinese
func T2S(s string) string {
	if chs, ok := t2sMapping[s]; ok {
		return chs
	}

	var buf bytes.Buffer
	for _, c := range s {
		v, ok := t2sMapping[string(c)]
		if ok {
			buf.WriteString(v)
		} else {
			buf.WriteString(string(c))
		}
	}
	return buf.String()
}

// Convert Simplified Chinese to Traditional Chinese
func S2T(s string) string {
	if cht, ok := s2tMapping[s]; ok {
		return cht
	}

	var buf bytes.Buffer
	for _, c := range s {
		v, ok := s2tMapping[string(c)]
		if ok {
			buf.WriteString(v)
		} else {
			buf.WriteString(string(c))
		}
	}
	return buf.String()
}
