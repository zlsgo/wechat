package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"github.com/sohaha/zlsgo/zstring"
)

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

func newECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func aesECBEncrypt(plaintext, key string) (ciphertext []byte, err error) {
	text := zstring.String2Bytes(plaintext)
	text = pkcs5Padding(text, aes.BlockSize)
	if len(text)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}
	aesKey := zstring.String2Bytes(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	cipher := make([]byte, len(text))
	newECBEncrypter(block).CryptBlocks(cipher, text)
	base64Msg := make([]byte, base64.StdEncoding.EncodedLen(len(cipher)))
	base64.StdEncoding.Encode(base64Msg, cipher)

	return base64Msg, nil
}

func aesECBDecrypt(ciphertext, key string) (plaintext []byte, err error) {
	text, _ := base64.StdEncoding.DecodeString(ciphertext)
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	if len(text)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	aesKey := zstring.String2Bytes(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	newECBDecrypter(block).CryptBlocks(text, text)

	plaintext = pkcs5UnPadding(text)
	return plaintext, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
