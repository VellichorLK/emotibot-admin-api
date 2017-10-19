package handlers

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
)

var key = []byte("isTjP-oq18KJZx73Trs905LlpF6kzmnO")

func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])
	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}
	return src[:(length - unpadding)], nil
}

func Encrypt(data []byte) ([]byte, error) {

	plaintext := Pad(data)
	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	if len(plaintext)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.

	return ciphertext, nil
}

func Decrpyt(data []byte) ([]byte, error) {

	ciphertext := data

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.

	// Output: exampleplaintext
	ciphertext, err = Unpad(ciphertext)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func shifter(data []byte) {
	var shift byte
	shift = 0x2c
	for i := 0; i < len(data); i += 2 {
		//if i%2 == 0 {
		data[i] = data[i] ^ shift
		//}
	}
}

func CreateCursor(byteInfo []byte) string {

	shifter(byteInfo)

	cipherText, err := Encrypt(byteInfo)
	if err != nil {
		log.Println(err)
		return ""
	}
	encodedText := base64.StdEncoding.EncodeToString(cipherText)
	return encodedText
	//return hex.EncodeToString(cipherText)
}

func DecryptCursor(cursor string) ([]byte, error) {
	decodedText, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	text, err := Decrpyt([]byte(decodedText))
	if err != nil {
		return nil, err
	}
	shifter(text)
	return text, nil
}

/*
func main() {

	//c := CreateCursor("t1=20170630")
	query := "t1=20170630&t2=2017063119&file_name=fuck.mp3&min_score=91"
	log.Printf("query:%s,%d", query, len(query))
	c := CreateCursor(query)

	d := DecryptCursor(c)

	log.Printf("src:%s, %d", d, len(d))

}
*/
