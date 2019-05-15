package fileservice

import (
	"bufio"
	"bytes"
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
			return fmt.Errorf("Check bucket - %s", err.Error())
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

func GetFile(namespace string, path string) ([]byte, error) {
	if minioClient == nil {
		return nil, ErrClientNotInit
	}

	if ok, err := minioClient.BucketExists(namespace); !ok {
		if err != nil {
			return nil, fmt.Errorf("Check bucket - %s", err.Error())
		}
		return nil, fmt.Errorf("Bucket not found")
	}

	obj, err := minioClient.GetObject(namespace, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("Get object - %s", err.Error())
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	_, err = io.Copy(writer, obj)
	if err != nil {
		return nil, fmt.Errorf("Buf copy - %s", err.Error())
	}

	return b.Bytes(), nil
}
