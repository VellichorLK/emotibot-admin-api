package util

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

const (
	// ModePerm is 777 for created dir shared with other docker
	ModePerm os.FileMode = 0777
)

func SaveDictionaryFile(appid string, filename string, file multipart.File) (int64, error) {
	dirPath := fmt.Sprintf("./Files/%s", appid)
	filePath := fmt.Sprintf("%s/%s", dirPath, filename)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		mkdirErr := os.MkdirAll(dirPath, ModePerm)
		if mkdirErr != nil {
			LogError.Printf("Cannot create appid dir into system (%s)", mkdirErr.Error())
			return 0, mkdirErr
		}
	}

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
	return written, nil
}
