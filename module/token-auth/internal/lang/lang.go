package lang

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var LG map[string]map[string]string

// 初始化语言包
func LoadLang() map[string]map[string]string {
	dir, err := filepath.Abs(filepath.Dir(os.Getenv("GOPATH") + "/src/emotibot.com/emotigo/module/token-auth/internal/lang"))
	if err != nil {

	}
	dir = dir + "/lang"

	dirs, err := ioutil.ReadDir(dir)
	if err != nil {

	}

	langDirs := map[string]string{}
	for _, fi := range dirs {
		if fi.IsDir() {
			langDirs[fi.Name()] = dir + "/" + fi.Name()
		}
	}

	LG = make(map[string]map[string]string)
	for local, v := range langDirs {
		lf, _ := ioutil.ReadDir(v)
		for _, vv := range lf {
			if filepath.Ext(vv.Name()) == ".json" {
				filePath := v + "/" + vv.Name()
				fileBytes, err := ioutil.ReadFile(filePath)
				if err != nil {

				}

				var mapLang map[string]string
				err = json.Unmarshal(fileBytes, &mapLang)
				if err != nil {

				}
				if _, ok := LG[local]; ok {
					for kkk, vvv := range mapLang {
						LG[local][kkk] = vvv
					}
				} else {
					LG[local] = mapLang
				}
			}
		}
	}
	//fmt.Println(LG)
	//jsonBytes, _ := json.Marshal(LG)
	//fmt.Println(string(jsonBytes))

	return LG
}

// 获取语言对应文字
func Get(local string, key string) string {
	if _, ok := LG[local]; !ok {
		local = "zh-cn"
	}
	if _, ok := LG[local][key]; !ok {
		return ""
	}

	return LG[local][key]
}
