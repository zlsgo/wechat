package wechat

import (
	"errors"
	"fmt"
	"time"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	Open struct {
		AppID           string
		AppSecret       string
		EncodingAesKey  string
		engine          *Engine
		refreshToken    string
		authorizerAppID string
		Token           string
	}
)

var (
	ErrOpenJumpAuthorization = errors.New(
		"need to jump to the authorization page")
)
var _ Cfg = new(Open)

func (o *Open) setEngine(engine *Engine) {
	o.engine = engine
}

func (o *Open) getEngine() *Engine {
	return o.engine
}

func (o *Open) GetSecret() string {
	return o.AppSecret
}
func (o *Open) GetToken() string {
	return o.Token
}

func (o *Open) GetEncodingAesKey() string {
	return o.EncodingAesKey
}

func (o *Open) checkEngine() (*Engine, error) {
	if o.engine == nil {
		return nil, errors.New(`please use wechat.New(&wechat.Open{})`)
	}
	return o.engine, nil
}

func (o *Open) GetComponentTicket() (string, error) {
	if _, err := o.checkEngine(); err != nil {
		return "", err
	}
	data, err := o.engine.cache.GetString(cacheComponentVerifyTicket)
	if err != nil {
		return "", errors.New("have not received wechat push information")
	}

	return data, nil
}

func (o *Open) GetComponentAccessToken() (string, error) {
	if _, err := o.checkEngine(); err != nil {
		return "", err
	}
	data, err := o.engine.cache.MustGet("component_access_token", func(set func(data interface{}, lifeSpan time.Duration, interval ...bool)) (err error) {
		var ticket string
		ticket, err = o.GetComponentTicket()
		if err != nil {
			return
		}
		post := zhttp.Param{
			"component_appid":         o.AppID,
			"component_appsecret":     o.AppSecret,
			"component_verify_ticket": ticket,
		}
		res, err := http.Post(fmt.Sprintf(
			"%s/cgi-bin/component/api_component_token", apiurl), zhttp.BodyJSON(post))
		if err != nil {
			return
		}
		var json *zjson.Res
		json, err = CheckResError(res.Bytes())
		if err != nil {
			return
		}
		componentAppsecret := json.Get("component_access_token").String()
		if componentAppsecret == "" {
			return errors.New("failed to parse component access token")
		}
		set(componentAppsecret, time.Duration(json.Get("expires_in").Int()-200)*time.Second)

		return
	})
	if err != nil {
		return "", err
	}
	return data.(string), nil

}

func (o *Open) getPreAuthCode() (string, error) {
	e, err := o.checkEngine()
	if err != nil {
		return "", err
	}
	var data interface{}
	data, err = e.cache.MustGet("pre_auth_code",
		func(set func(data interface{}, lifeSpan time.Duration, interval ...bool)) (err error) {
			var ticket string
			ticket, err = o.GetComponentAccessToken()
			if err != nil {
				return
			}
			url := fmt.Sprintf("%s/cgi-bin/component/api_create_preauthcode?component_access_token=%s", apiurl, ticket)
			var res *zhttp.Res
			post, _ := zjson.Set("{}", "component_appid", o.AppID)
			res, err = http.Post(url, post)
			if err != nil {
				return
			}
			var json *zjson.Res
			json, err = CheckResError(res.Bytes())
			if err != nil {
				return
			}
			authCode := json.Get("pre_auth_code").String()
			set(authCode, time.Duration(json.Get("expires_in").Int()-200)*time.Second)

			return
		})

	return data.(string), nil

}

func (e *Engine) GetConfig() Cfg {
	return e.config
}

// ComponentVerifyTicket 解析微信开放平台 Ticket
func (e *Engine) ComponentVerifyTicket(raw string) (
	string, error) {
	if !e.IsOpen() {
		return "", errors.New("only supports open")
	}
	config, ok := e.config.(*Open)
	if !ok {
		return "", errors.New("only supports open")
	}
	data, _ := ParseXML2Map(zstring.String2Bytes(raw))
	Encrypt, ok := data["Encrypt"]
	if !ok {
		return "", errors.New("illegal data")
	}
	var (
		err        error
		cipherText []byte
	)
	cipherText, err = aesDecrypt(Encrypt, config.EncodingAesKey)
	appidOffset := len(cipherText) - len(zstring.String2Bytes(config.AppID))
	if appid := string(cipherText[appidOffset:]); appid != config.AppID {
		return "", errors.New("appid mismatch")
	}
	cipherText = cipherText[20:appidOffset]
	var ticketData XMLData
	ticketData, err = ParseXML2Map(cipherText)
	if err != nil {
		return "", err
	}
	ticket := ticketData["ComponentVerifyTicket"]
	log.Debug("收到 Ticket:", ticket)
	e.cache.Set(cacheComponentVerifyTicket, ticket, 0)
	return ticket, nil
}

func (o *Open) GetAppID() string {
	return o.AppID
}

func (e *Engine) ComponentApiQueryAuth(authCode, redirectUri string) (s string,
	redirect string, err error) {
	if !e.IsOpen() {
		return "", "", errors.New("only supports open")
	}
	config := e.config.(*Open)
	if authCode == "" {
		return e.getAuthUri(config, redirectUri)
	}
	componentAccessToken, err := config.GetComponentAccessToken()
	if err != nil {
		return "", "", err
	}
	res, err := http.Post(fmt.Sprintf(
		"%s/cgi-bin/component/api_query_auth?component_access_token=%s", apiurl,
		componentAccessToken), zhttp.BodyJSON(
		map[string]string{
			"component_appid":    e.GetAppId(),
			"authorization_code": authCode,
		}))
	if err != nil {
		return "", "", err
	}
	json, err := CheckResError(res.Bytes())
	if err != nil {
		return e.getAuthUri(config, redirectUri)
	}
	return json.String(), "", nil
}

func (e *Engine) getAuthUri(config *Open, redirectUri string) (string, string, error) {
	preAuthCode, err := config.getPreAuthCode()
	if err != nil {
		return "", "", err
	}
	url := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s", e.GetAppId(), preAuthCode, redirectUri)
	return "", url, ErrOpenJumpAuthorization
}

func (o *Open) getAccessToken() (data []byte, err error) {
	if o.refreshToken == "" {
		err = errors.New("please authorize it through the ComponentApiQueryAuth method")
		return
	}
	var componentAccessToken string
	componentAccessToken, err = o.GetComponentAccessToken()
	if err != nil {
		return
	}
	res, err := http.Post(fmt.Sprintf(
		"%s/cgi-bin/component/api_authorizer_token?component_access_token=%s", apiurl, componentAccessToken), zhttp.BodyJSON(zhttp.Param{
		"component_appid":          o.AppID,
		"authorizer_appid":         o.authorizerAppID,
		"authorizer_refresh_token": o.refreshToken,
	}))
	if err != nil {
		return
	}
	var json *zjson.Res
	json, err = CheckResError(res.Bytes())
	if err != nil {
		return
	}
	refreshToken := json.Get("authorizer_refresh_token").String()
	if refreshToken == "" {
		err = errors.New("failed to parse api authorizer token")
		return
	}
	o.refreshToken = refreshToken

	return res.Bytes(), nil
}

func (o *Open) SetAuthorizerAccessToken(authorizerAppID, accessToken,
	refreshToken string, expiresIn uint) {
	o.refreshToken = refreshToken
	o.authorizerAppID = authorizerAppID
	o.engine.cache.Set(cacheToken, accessToken, expiresIn)
}

func (o *Open) getJsapiTicket() (data *zhttp.Res, err error) {
	var token string
	token, err = o.engine.GetAccessToken()
	if err != nil {
		return nil, err
	}
	return http.Post(fmt.Sprintf(
		"%s/cgi-bin/ticket/getticket?&type=jsapi&access_token=%s",
		apiurl, token))
}
