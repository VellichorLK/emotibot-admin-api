package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

const (
	// ModePerm is 777 for created dir shared with other docker
	ModePerm        os.FileMode = 0777
	mountDirPathKey string      = "MOUNT_PATH"
)

// GetFunctionSettingPath will return <appid>.property path of appid
func GetFunctionSettingPath(appid string) string {
	appidDir, _ := getAppidDir(appid)
	configPath := fmt.Sprintf("%s/%s.property", appidDir, appid)
	return configPath
}

func GetFunctionSettingTmpPath(appid string) string {
	appidDir, _ := getAppidDir(appid)
	configPath := fmt.Sprintf("%s/%s_function_tmp.property", appidDir, appid)
	return configPath
}

func SaveDictionaryFile(appid string, filename string, file multipart.File) (int64, error) {
	dirPath, err := getAppidDir(appid)
	if err != nil {
		return 0, err
	}
	filePath := fmt.Sprintf("%s/%s", dirPath, filename)
	output, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		LogError.Printf("Cannot create file (%s)", err.Error())
		return 0, err
	}
	defer output.Close()

	written, err := io.Copy(output, file)
	if err != nil {
		LogError.Printf("Cannot copy file into system (%s)", err.Error())
		return 0, err
	}

	LogTrace.Printf("Write to file [%s] [%d] bytes successfully", filePath, written)
	return written, nil
}

func getAppidDir(appid string) (string, error) {
	mountPath := getGlobalEnv(mountDirPathKey)
	LogTrace.Printf("Mount path: %s", mountPath)

	dirPath := fmt.Sprintf("%s/settings/%s", mountPath, appid)
	mkdirErr := os.MkdirAll(dirPath, ModePerm)
	if mkdirErr != nil {
		LogError.Printf("Cannot create appid dir into system (%s)", mkdirErr.Error())
	}
	return dirPath, mkdirErr
}

func GetMountDir() string {
	mountPath := getGlobalEnv(mountDirPathKey)
	LogTrace.Printf("Mount path: %s", mountPath)
	return mountPath
}

func SaveNLUFileFromEntity(appid string, wordLines []string, synonyms []string) (err error, md5Words string, md5Synonyms string) {
	dirPath, err := getAppidDir(appid)
	if err != nil {
		return
	}
	wordsFileName := fmt.Sprintf("%s.txt", appid)
	synonymsFileName := fmt.Sprintf("%s_synonyms.txt", appid)

	wordsFilePath := fmt.Sprintf("%s/%s", dirPath, wordsFileName)
	synonymsFilePath := fmt.Sprintf("%s/%s", dirPath, synonymsFileName)

	outputWords, err := os.OpenFile(wordsFilePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		LogError.Printf("Cannot create file (%s)", err.Error())
		return
	}
	defer outputWords.Close()

	outWords := strings.Join(wordLines, "\n") + "\n"
	md5Words = fmt.Sprintf("%x", md5.Sum([]byte(outWords)))
	_, err = outputWords.WriteString(outWords)
	if err != nil {
		return
	}

	outputSynonyms, err := os.OpenFile(synonymsFilePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		LogError.Printf("Cannot create file (%s)", err.Error())
		return
	}
	defer outputSynonyms.Close()

	outSynonyms := strings.Join(synonyms, "\n") + "\n"
	md5Synonyms = fmt.Sprintf("%x", md5.Sum([]byte(outSynonyms)))
	_, err = outputSynonyms.WriteString(outSynonyms)
	if err != nil {
		return
	}
	return
}