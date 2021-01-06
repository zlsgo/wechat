package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/sohaha/zlsgo/zstring"
)

func sha1Signature(params ...string) string {
	sort.Strings(params)
	h := sha1.New()
	for _, s := range params {
		_, _ = io.WriteString(h, s)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func aesEncrypt(data string, key string, iv ...byte) ([]byte, error) {
	aesKey := encodingAESKey2AESKey(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	if len(iv) == 0 {
		iv = aesKey[:block.BlockSize()]
	}
	plainText := PKCS7Padding(zstring.String2Bytes(data), len(aesKey))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	base64Msg := make([]byte, base64.StdEncoding.EncodedLen(len(cipherText)))
	base64.StdEncoding.Encode(base64Msg, cipherText)

	return base64Msg, nil
}

func aesDecrypt(data string, key string, iv ...string) ([]byte,
	error) {
	aesKey := encodingAESKey2AESKey(key)
	cipherData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	var ivRaw []byte
	plainText := make([]byte, len(cipherData))
	if len(iv) == 0 {
		ivRaw = aesKey[:block.BlockSize()]
	} else {
		ivRaw = zstring.String2Bytes(iv[0])
	}
	blockMode := cipher.NewCBCDecrypter(block, ivRaw)
	blockMode.CryptBlocks(plainText, cipherData)

	return PKCS7UnPadding(plainText, len(key)), nil
}

func encodingAESKey2AESKey(encodingKey string) []byte {
	data, _ := base64.StdEncoding.DecodeString(encodingKey + "=")
	return data
}

func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	if padding == 0 {
		padding = blockSize
	}
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS7UnPadding(plainText []byte, blockSize int) []byte {
	l := len(plainText)
	unpadding := int(plainText[l-1])
	if unpadding < 0 || unpadding > blockSize {
		unpadding = 0
	}
	return plainText[:(l - unpadding)]
}

func parsePlainText(plaintext []byte) ([]byte, uint32, []byte, []byte, error) {
	textLen := uint32(len(plaintext))
	if textLen < 20 {
		return nil, 0, nil, nil, errors.New("plain is to small 1")
	}
	random := plaintext[:16]
	msgLen := binary.BigEndian.Uint32(plaintext[16:20])
	if textLen < (20 + msgLen) {
		return nil, 0, nil, nil, errors.New("plain is to small 2")
	}
	msg := plaintext[20 : 20+msgLen]
	receiverId := plaintext[20+msgLen:]
	return random, msgLen, msg, receiverId, nil
}

func MarshalPlainText(replyMsg, receiverId, random string) string {
	var buffer bytes.Buffer
	buffer.WriteString(random)
	msgLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBuf, uint32(len(replyMsg)))
	buffer.Write(msgLenBuf)
	buffer.WriteString(replyMsg)
	buffer.WriteString(receiverId)
	return buffer.String()
}
