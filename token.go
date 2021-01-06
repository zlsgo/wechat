package wechat

import (
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zjson"
)

// AccessToken accessToken
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (e *Engine) GetAccessTokenExpiresInCountdown() float64 {
	data, err := e.cache.GetT(cacheToken)
	if err != nil {
		return 0
	}
	return data.RemainingLife().Seconds()
}

// 设置 AccessToken
func (e *Engine) SetAccessToken(accessToken string, expiresIn uint) error {
	e.cache.Set(cacheToken, accessToken, expiresIn-60)
	return nil
}

// 获取 AccessToken
func (e *Engine) GetAccessToken() (string, error) {
	data, err := e.cache.MustGet(cacheToken, func(set func(data interface{},
		lifeSpan time.Duration, interval ...bool)) (err error) {
		var res []byte
		res, err = e.config.getAccessToken()
		if err != nil {
			return
		}
		var json *zjson.Res
		json, err = CheckResError(res)
		if err != nil {
			return err
		}
		accessToken := json.Get("access_token").String()
		if accessToken == "" {
			accessToken = json.Get("authorizer_access_token").String()
		}
		if accessToken == "" {
			return errors.New("access_token parsing failed")
		}
		set(accessToken, time.Duration(json.Get("expires_in").Int()-200)*time.Second)
		return nil
	})
	if err != nil {
		return "", err
	}

	return data.(string), nil
}
