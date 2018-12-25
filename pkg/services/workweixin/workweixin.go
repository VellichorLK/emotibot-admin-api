package workweixin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type Client struct {
	Token       string
	EncodingAES string
	AESKey      []byte
	Cipher      cipher.Block
}

func New(token, encodingAES string) (*Client, error) {
	client := &Client{}
	client.Token = token
	client.EncodingAES = encodingAES

	var err error
	client.AESKey, err = base64.StdEncoding.DecodeString(encodingAES)
	if err != nil {
		return nil, err
	}

	client.Cipher, err = aes.NewCipher(client.AESKey)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) ParseRequest(r *http.Request) {

}

func (c *Client) VerifyURL(w http.ResponseWriter, r *http.Request) {
	signature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	encryptStr := r.URL.Query().Get("echostr")

	logger.Trace.Printf("Verify with: %s, %s, %s, %s\n", signature, timestamp, nonce, encryptStr)
	verify := calculateSignature(c.Token, timestamp, nonce, encryptStr)
	logger.Trace.Printf("Signature check: %s, %s\n", verify, signature)
	if verify != signature {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decodedStr, err := base64.StdEncoding.DecodeString(encryptStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("Verify URL error when decode base64, ", err.Error())
		return
	}
	outputStr := make([]byte, len(decodedStr))

	c.Cipher.Decrypt(outputStr, decodedStr)
	content := outputStr[16:]
	msgLen, err := strconv.Atoi(string(outputStr[0:4]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Printf("Verify URL error when get msgLen %s, %s\n", outputStr[0:4], err.Error())
		return
	}
	msg := content[4 : msgLen+4]
	verified := content[msgLen+4:]

	logger.Trace.Printf("Get msg [%s], verified [%s]\n", msg, verified)
	w.Write(msg)
}

func calculateSignature(token, timestamp, nonce, message string) string {
	params := []string{token, timestamp, nonce, message}
	sort.Strings(params)
	input := strings.Join(params, "")
	logger.Trace.Printf("Sorted strings: %s\n", input)
	hash := sha1.New()
	io.WriteString(hash, input)
	signature := fmt.Sprintf("%x", hash.Sum(nil))
	return signature
}
