package fileservice

import (
	"errors"
	"fmt"
	"io"

	minio "github.com/minio/minio-go"
)

var (
	minioClient *minio.Client

	ErrClientNotInit = errors.New("minio client not init")
)

func Init(url, accessKey, secretKey string, useSSL bool) error {
	var err error
	minioClient, err = minio.New(url, accessKey, secretKey, useSSL)
	if err != nil {
		return err
	}
	return nil
}

func AddFile(namespace string, path string, contentReader io.Reader) error {
	if minioClient == nil {
		return ErrClientNotInit
	}

	if ok, err := minioClient.BucketExists(namespace); !ok {
		if err != nil {
			return fmt.Errorf("Create bucket - %s", err.Error())
		}
		err = minioClient.MakeBucket(namespace, "")
		if err != nil {
			return fmt.Errorf("Make bucket - %s", err.Error())
		}
	}

	_, err := minioClient.PutObject(namespace, path, contentReader, -1, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("Put object - %s", err.Error())
	}

	return nil
}
