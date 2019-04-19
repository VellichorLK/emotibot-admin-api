package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"encoding/base64"
)

const (
	AesKey     = "emotibot@airobot"
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}





func AesBase64Encrypt(data string)(string, error){
	var aeskey = []byte(AesKey)
	value := []byte(data)
	xpass, err := AesEncrypt(value, aeskey)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	base64Value := base64.StdEncoding.EncodeToString(xpass)
	fmt.Printf("after AesEncrypt :%v\n",base64Value)

	return base64Value, err
}

func AesBase64Decrypt(data string)(string, error) {
	var aeskey = []byte(AesKey)
	base64Value, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	value, err := AesDecrypt(base64Value, aeskey)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Printf("after AesDecrypt :%s\n", value)
	return string(value), err
}



//func main() {
//	var aeskey = []byte("321423u9y8d2fwfl")
//	pass := []byte("vdncloud123456")
//	xpass, err := AesEncrypt(pass, aeskey)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	pass64 := base64.StdEncoding.EncodeToString(xpass)
//	fmt.Printf("加密后:%v\n",pass64)
//
//	bytesPass, err := base64.StdEncoding.DecodeString(pass64)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	tpass, err := AesDecrypt(bytesPass, aeskey)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Printf("解密后:%s\n", tpass)
//}

