package wechat

import (
	"fmt"

	"github.com/sohaha/zlsgo/zhttp"
)

type (
	Qy struct {
		CorpID         string
		Secret         string
		EncodingAesKey string
		engine         *Engine
		Token          string
	}
)

var _ Cfg = new(Qy)

func (q *Qy) setEngine(engine *Engine) {
	q.engine = engine
}

func (q *Qy) getEngine() *Engine {
	return q.engine
}

func (q *Qy) GetAppID() string {
	return q.CorpID
}

func (q *Qy) GetSecret() string {
	return q.Secret
}

func (q *Qy) GetToken() string {
	return q.Token
}

func (q *Qy) GetEncodingAesKey() string {
	return q.EncodingAesKey
}

func (q *Qy) getAccessToken() (data []byte, err error) {
	var res *zhttp.Res
	res, err = http.Post(fmt.Sprintf(
		"%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", qyurl, q.CorpID,
		q.Secret))
	if err != nil {
		return
	}
	data = res.Bytes()
	return
}

func (q *Qy) getJsapiTicket() (data *zhttp.Res, err error) {
	var token string
	token, err = q.engine.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return http.Post(fmt.Sprintf(
		"%s/cgi-bin/get_jsapi_ticket?access_token=%s",
		qyurl, token))
}
