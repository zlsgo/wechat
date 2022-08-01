package wechat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	Weapp struct {
		AppID          string
		AppSecret      string
		EncodingAesKey string
		Token          string
		engine         *Engine
	}
)

var _ Cfg = new(Weapp)

func (m *Weapp) setEngine(engine *Engine) {
	m.engine = engine
}

func (m *Weapp) getEngine() *Engine {
	return m.engine
}

func (m *Weapp) GetAppID() string {
	return m.AppID
}

func (m *Weapp) GetSecret() string {
	return m.AppSecret
}

func (m *Weapp) GetToken() string {
	return m.Token
}

func (m *Weapp) GetEncodingAesKey() string {
	return m.EncodingAesKey
}

func (m *Weapp) getAccessToken() (data []byte, err error) {
	res, err := http.Post(fmt.Sprintf(
		"%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		APIURL, m.AppID, m.AppSecret))
	if err != nil {
		return
	}
	data = res.Bytes()
	return
}

func (m *Weapp) getJsapiTicket() (data *zhttp.Res, err error) {
	var token string
	token, err = m.engine.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return http.Post(fmt.Sprintf(
		"%s/cgi-bin/ticket/getticket?&type=jsapi&access_token=%s",
		APIURL, token))
}

func (m *Weapp) GetSessionKey(code, grantType string) (data *zjson.Res, err error) {
	u := zstring.Buffer(9)
	u.WriteString(APIURL)
	u.WriteString("/sns/jscode2session?appid=")
	u.WriteString(m.AppID)
	u.WriteString("&secret=")
	u.WriteString(m.AppSecret)
	u.WriteString("&js_code=")
	u.WriteString(code)
	u.WriteString("&grant_type=")
	u.WriteString(grantType)
	return httpResProcess(http.Get(u.String()))
}

func (m *Weapp) Decrypt(seesionKey, iv, encryptedData string) (string, error) {
	byts := make([][]byte, 0, 3)
	for _, v := range []string{seesionKey, iv, encryptedData} {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return "", err
		}
		byts = append(byts, b)
	}
	aesKey := byts[0]
	ivRaw := byts[1]
	cipherData := byts[2]
	block, _ := aes.NewCipher(aesKey)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, ivRaw)
	plaintext := make([]byte, len(cipherData))
	blockMode.CryptBlocks(plaintext, cipherData)
	plaintext = PKCS7UnPadding(plaintext, blockSize)
	return zstring.Bytes2String(plaintext), nil
}

func (m *Weapp) Verify(seesionKey, rawData, signature string) bool {
	h := sha1.New()
	h.Write(zstring.String2Bytes(rawData))
	h.Write(zstring.String2Bytes(seesionKey))
	sign := fmt.Sprintf("%x", h.Sum(nil))
	return sign == signature
}
