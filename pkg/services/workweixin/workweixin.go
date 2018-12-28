package workweixin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
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

func (c *Client) VerifyURL(w http.ResponseWriter, r *http.Request) {
	signature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	encryptStr := r.URL.Query().Get("echostr")

	logger.Trace.Printf(`Verify with:
	signature:	%s
	timestamp: %s
	nonce: %s
	encryptStr: %s
	`, signature, timestamp, nonce, encryptStr)
	verify := calculateSignature(c.Token, timestamp, nonce, encryptStr)
	logger.Trace.Printf("Signature check: %s, %s\n", verify, signature)
	if verify != signature {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msg, _, err := decrypt(c.Cipher, c.AESKey, encryptStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Printf("Decrypt error: %s\n", err.Error())
		return
	}
	w.Write(msg)
}

func decrypt(c cipher.Block, key []byte, encryptStr string) ([]byte, []byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(encryptStr)
	if err != nil {
		return nil, nil, err
	}

	blockMode := cipher.NewCBCDecrypter(c, key[0:16])
	outputStr := make([]byte, len(decodedStr))
	blockMode.CryptBlocks(outputStr, decodedStr)
	outputStr = PKCS5UnPadding(outputStr)
	content := outputStr[16:]
	msgLen := binary.BigEndian.Uint32(content[:4])
	if int(msgLen) > len(content) {
		logger.Error.Printf("length too large")
		return nil, nil, err
	}
	msg := content[4 : msgLen+4]
	verified := content[msgLen+4:]
	logger.Trace.Printf("Get origin result: %s\n", content)
	logger.Trace.Printf("Get msg %s\n", msg)
	logger.Trace.Printf("Get verified: %s\n", verified)
	return msg, verified, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
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

func (c *Client) ParseRequest(r *http.Request) ([]*WorkWeixinMessage, error) {
	_, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		return nil, nil
	}
	return nil, nil
}
