package util

import (
	"archive/zip"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util/localemsg"

	"emotibot.com/emotigo/pkg/logger"
)

const (
	// ModePerm is 777 for created dir shared with other docker
	ModePerm             os.FileMode = 0777
	mountDirPathKey      string      = "MOUNT_PATH"
	wordbankTemplateFile string      = "wordbank_template.xlsx"
)

// GetCurDir returns the current directory of the executable binary
func GetCurDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(exe), nil
}

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
		logger.Error.Printf("Cannot create file (%s)", err.Error())
		return 0, err
	}
	defer output.Close()

	written, err := io.Copy(output, file)
	if err != nil {
		logger.Error.Printf("Cannot copy file into system (%s)", err.Error())
		return 0, err
	}

	logger.Trace.Printf("Write to file [%s] [%d] bytes successfully", filePath, written)
	return written, nil
}

func getAppidDir(appid string) (string, error) {
	mountPath := getGlobalEnv(mountDirPathKey)
	logger.Trace.Printf("Mount path: %s", mountPath)

	dirPath := fmt.Sprintf("%s/settings/%s", mountPath, appid)
	mkdirErr := os.MkdirAll(dirPath, ModePerm)
	if mkdirErr != nil {
		logger.Error.Printf("Cannot create appid dir into system (%s)", mkdirErr.Error())
	}
	return dirPath, mkdirErr
}

func GetMountDir() string {
	mountPath := getGlobalEnv(mountDirPathKey)
	logger.Trace.Printf("Mount path: %s", mountPath)
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
		logger.Error.Printf("Cannot create file (%s)", err.Error())
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
		logger.Error.Printf("Cannot create file (%s)", err.Error())
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

func GetWordbankTemplatePath(locale string) string {
	templateDir := GetTemplateDir(locale)
	return fmt.Sprintf("%s/%s", templateDir, wordbankTemplateFile)
}

func CompressFiles(filePaths []string, outFilePath string) error {
	zipFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, fPath := range filePaths {
		file, err := os.Open(fPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get the file information
		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use deflate compression
		header.Method = zip.Deflate

		// Compress the file
		w, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, file); err != nil {
			return err
		}
	}

	return nil
}

func GetTemplateDir(locale string) string {
	actualLocale := locale
	if locale != localemsg.ZhCn && locale != localemsg.ZhTw {
		actualLocale = localemsg.ZhCn
	}
	mountPath := getGlobalEnv(mountDirPathKey)
	return fmt.Sprintf("%s/%s", mountPath, actualLocale)
}
