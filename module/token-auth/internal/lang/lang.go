package lang

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"emotibot.com/emotigo/module/token-auth/internal/util"
)

var LG map[string]map[string]string

// 初始化语言包
func LoadLang() map[string]map[string]string {
	dir := "template/lang"

	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		util.LogInfo.Println("lang cannot read dir")
		return nil
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
					util.LogInfo.Println("lang read file failed")
					return nil
				}

				var mapLang map[string]string
				err = json.Unmarshal(fileBytes, &mapLang)
				if err != nil {
					util.LogInfo.Println("lang unmarshal json failed")
					return nil
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

	return LG
}

// 获取语言对应文字
func Get(locale string, key string) string {
	if _, ok := LG[locale]; !ok {
		util.LogInfo.Println("lang get no locale")
		locale = "zh-cn"
	}
	if _, ok := LG[locale][key]; !ok {
		util.LogInfo.Println("lang get no key")
		return ""
	}

	return LG[locale][key]
}
