package wechat

import (
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
