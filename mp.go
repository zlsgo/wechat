package wechat

import (
	"fmt"

	"github.com/sohaha/zlsgo/zhttp"
)

type (
	Mp struct {
		// 公众号 ID
		AppID string
		// 公众号密钥
		AppSecret      string
		EncodingAesKey string
		Token          string
		engine         *Engine
	}
)

var _ Cfg = new(Mp)

func (m *Mp) setEngine(engine *Engine) {
	m.engine = engine
}

func (m *Mp) getEngine() *Engine {
	return m.engine
}

func (m *Mp) GetAppID() string {
	return m.AppID
}

func (m *Mp) GetSecret() string {
	return m.AppSecret
}

func (m *Mp) GetToken() string {
	return m.Token
}

func (m *Mp) GetEncodingAesKey() string {
	return m.EncodingAesKey
}

func (m *Mp) getAccessToken() (data []byte, err error) {
	res, err := http.Post(fmt.Sprintf(
		"%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		apiurl, m.AppID, m.AppSecret))
	if err != nil {
		return
	}
	data = res.Bytes()
	return
}

func (m *Mp) getJsapiTicket() (data *zhttp.Res, err error) {
	var token string
	token, err = m.engine.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return http.Post(fmt.Sprintf(
		"%s/cgi-bin/ticket/getticket?&type=jsapi&access_token=%s",
		apiurl, token))
}
