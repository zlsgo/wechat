package wechat

import (
	"fmt"
	"time"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	JsSign struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
		NonceStr  string `json:"nonce_str"`
		Signature string `json:"signature"`
	}
)

func (e *Engine) GetJsSign(url string) (JsSign, error) {
	jsapiTicket, err := e.GetJsapiTicket()
	if err != nil {
		return JsSign{}, err
	}
	timestamp := time.Now().Unix()
	noncestr := zstring.Rand(16)
	signature := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", jsapiTicket, noncestr, timestamp, url)
	signature = sha1Signature(signature)
	return JsSign{
		AppID:     e.GetAppId(),
		NonceStr:  noncestr,
		Timestamp: timestamp,
		Signature: signature,
	}, nil
}

func (e *Engine) SetJsapiTicket(ticket string, expiresIn uint) error {
	e.cache.Set(cacheJsapiTicket, ticket, expiresIn-60)
	return nil
}

func (e *Engine) GetJsapiTicket() (string, error) {
	data, err := e.cache.MustGet(cacheJsapiTicket, func(set func(data interface{}, lifeSpan time.Duration, interval ...bool)) (err error) {
		var res *zhttp.Res
		res, err = e.config.getJsapiTicket()
		if err != nil {
			return
		}
		var json *zjson.Res
		json, err = CheckResError(res.Bytes())
		if err != nil {
			return
		}
		ticket := json.Get("ticket").String()
		if ticket == "" {
			return
		}
		set(ticket, time.Duration(json.Get("expires_in").Int()-200)*time.Second)
		return
	})
	if err != nil {
		return "", err
	}
	return data.(string), nil

}
