package workweixin

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

type Client struct {
	Token       string
	EncodingAES string
	AESKey      []byte
	Cipher      cipher.Block

	CorpID      string
	Secret      string
	AccessToken string
	ExpireTime  int64
}

func New(corpid, secret, token, encodingAES string) (*Client, error) {
	client := &Client{}
	client.Token = token
	client.EncodingAES = encodingAES
	client.Secret = secret
	client.CorpID = corpid

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

// VerifyURL will used to reply workweixin validation
// refs: https://work.weixin.qq.com/api/doc#90000/90135/90237/%E9%AA%8C%E8%AF%81URL%E6%9C%89%E6%95%88%E6%80%A7
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

	if strings.Trim(verify, " ") != strings.Trim(signature, " ") {
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

// decrypt will do the decryption of work weixin
// refs: https://work.weixin.qq.com/api/doc#90000/90139/90968/%E5%AF%86%E6%96%87%E8%A7%A3%E5%AF%86%E5%BE%97%E5%88%B0msg%E7%9A%84%E8%BF%87%E7%A8%8B
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

// PKCS5UnPadding will undo the padding used by workweixin
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// calculateSignature will do the signature of work weixin to check API call validation
// refs: https://work.weixin.qq.com/api/doc#90000/90139/90968/%E6%B6%88%E6%81%AF%E4%BD%93%E7%AD%BE%E5%90%8D%E6%A0%A1%E9%AA%8C
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

// getPostMsg will do the decrypt and return request body
func (c *Client) getPostMsg(r *http.Request) ([]byte, error) {
	content, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	input := Input{}
	err = xml.Unmarshal(content, &input)
	if err != nil {
		return nil, err
	}
	signature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	encryptStr := input.Encrypted

	logger.Trace.Printf(`Verify with:
	signature:	%s
	timestamp: %s
	nonce: %s
	encryptStr: %s
	`, signature, timestamp, nonce, encryptStr)
	verify := calculateSignature(c.Token, timestamp, nonce, encryptStr)
	logger.Trace.Printf("Signature check: %s, %s\n", verify, signature)
	if verify != signature {
		return nil, ErrInvalidSignature
	}

	msg, _, err := decrypt(c.Cipher, c.AESKey, encryptStr)
	if err != nil {
		logger.Error.Printf("Decrypt error: %s\n", err.Error())
		return nil, err
	}
	return msg, nil
}

// ParseRequest will parse the request to return the actual work weixin message
func (c *Client) ParseRequest(r *http.Request) (Message, error) {
	input, err := c.getPostMsg(r)
	if err != nil {
		return nil, err
	}

	rawMsg := rawMessage{}
	err = xml.Unmarshal(input, &rawMsg)
	if err != nil {
		return nil, err
	}

	switch rawMsg.Type {
	case MessageTypeText:
		textMsg := TextMessage{}
		err = xml.Unmarshal(input, &textMsg)
		if err != nil {
			return nil, err
		}
		logger.Trace.Printf("Receive %+v\n", textMsg)
		return &textMsg, nil
	case MessageTypeImage:

	}

	return nil, nil
}

// NewTextMessage will create new text message node to send to user
func NewTextMessage(receiver string, agentID int, text string) SendingMessage {
	ret := TextSendMessage{}
	ret.To = receiver
	ret.Type = MessageTypeText
	ret.AgentID = agentID
	ret.Text = &TextNode{text}
	return &ret
}

// SendMessages will send message to user
func (c *Client) SendMessages(messages []SendingMessage) error {
	errMsg := ""
	for idx := range messages {
		logger.Trace.Println("Send message, ", messages[idx])
		if messages[idx] == nil {
			continue
		}
		input, err := json.Marshal(messages[idx])
		if err != nil {
			logger.Error.Println("Marshal json fail:", err.Error())
			return err
		}
		ret, err := c.Post(MsgSendURL, input)
		logger.Trace.Println("Get reply:", ret)
		if err != nil {
			errMsg += err.Error()
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

// Post will post to work weixin with valid token
func (c *Client) Post(url string, input []byte) (*APIChatReturn, error) {
	var err error
	if !c.AccessTokenValidate() {
		err = c.GetNewAccessToken()
		if err != nil {
			return nil, err
		}
	}

	reader := bytes.NewReader(input)
	realURL := fmt.Sprintf("%s?access_token=%s", url, c.AccessToken)
	logger.Trace.Printf("Send request to: %s, with body: %s\n", realURL, input)

	rsp, err := http.Post(realURL, "application/json", reader)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	logger.Trace.Printf("Post get: %s\n", body)
	ret := APIChatReturn{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// AccessTokenValidate will check stored token is expired or not
func (c *Client) AccessTokenValidate() bool {
	// Commented method is a hacked method to do the validation
	// if c.AccessToken == "" {
	// 	return false
	// }
	// rsp, err := http.Get(fmt.Sprintf("%s?access_token=%s", TokenValidateURL, c.AccessToken))
	// if err != nil {
	// 	logger.Error.Println("Check access token validation fail,", err.Error())
	// 	return false
	// }
	// defer rsp.Body.Close()
	// content, err := ioutil.ReadAll(rsp.Body)
	// if err != nil {
	// 	logger.Error.Println("Read body error,", err.Error())
	// 	return false
	// }
	// if bytes.Contains(content, []byte("data")) {
	// 	return true
	// }
	// return false
	now := time.Now()
	return now.Unix() < c.ExpireTime
}

// GetNewAccessToken will use work weixin API to issue a new token
// refs: https://work.weixin.qq.com/api/doc#90000/90135/91039
func (c *Client) GetNewAccessToken() error {
	now := time.Now()
	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", TokenIssueURL, c.CorpID, c.Secret)
	logger.Trace.Println("Get token with url:", url)
	rsp, err := http.Get(url)
	if err != nil {
		logger.Error.Println("Get new access token request fail,", err.Error())
		return err
	}
	defer rsp.Body.Close()

	decoder := json.NewDecoder(rsp.Body)
	ret := APIAccessTokenReturn{}
	err = decoder.Decode(&ret)
	if err != nil {
		logger.Error.Println("Decode return fail:", err.Error())
		return err
	}

	c.AccessToken = ret.AccessToken
	// minus 10 second to avoid API latency
	c.ExpireTime = ret.Expire + now.Unix() - 10

	logger.Trace.Printf("Get new access token: %s, %d\n", c.AccessToken, c.ExpireTime)
	return nil
}
