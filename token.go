package wechat

import (
	"errors"
	"net/url"
	"time"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

// AccessToken accessToken
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ScopeType string

const (
	ScopeBase     ScopeType = "snsapi_base"
	ScopeUserinfo ScopeType = "snsapi_userinfo"
	// ScopePrivateinfo 企业微信需要使用这个才能拿到用户的基本信息
	ScopePrivateinfo ScopeType = "snsapi_privateinfo"
)

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

// Auth 用户授权
func (e *Engine) Auth(c *znet.Context, state string, scope ScopeType) (*zjson.Res, bool, error) {
	code := e.authCode(c, state, scope, "", "")
	if len(code) == 0 {
		return nil, false, nil
	}
	json, err := e.GetAuthInfo(code)
	if err != nil {
		if httpErr, ok := err.(httpError); ok {
			switch httpErr.Code {
			case 41008, 40029, 40163:
				if len(e.authCode(c, state, scope, "", code)) == 0 {
					return nil, false, nil
				}
			}
		}
	}
	return json, true, err
}

func (e *Engine) authCode(c *znet.Context, state string, scope ScopeType, uri, oldCode string) string {
	code, _ := c.GetQuery("code")
	if len(code) > 0 && code != oldCode {
		return code
	}

	if len(uri) == 0 {
		if len(e.redirectDomain) > 0 {
			uri = e.redirectDomain + c.Request.URL.String()
		} else {
			uri = c.Host(true)
		}
	}

	c.Redirect(e.getOauthRedirect(paramFilter(uri), state, scope))
	c.Abort()
	return ""
}

func (e *Engine) getOauthRedirect(callback string, state string, scope ScopeType) string {
	if len(scope) == 0 {
		scope = "snsapi_userinfo"
	}
	u := zstring.Buffer(10)
	if e.IsQy() {
		u.WriteString("https://open.weixin.qq.com/connect/oauth2/authorize?appid=")
		u.WriteString(e.GetAppID())
		u.WriteString("&agentid=")
		conf := e.config.(*Qy)
		u.WriteString(conf.AgentID)
	} else {
		u.WriteString(openURL)
		u.WriteString("/connect/oauth2/authorize?appid=")
		u.WriteString(e.GetAppID())
	}

	u.WriteString("&redirect_uri=")
	u.WriteString(url.QueryEscape(callback))
	u.WriteString("&response_type=code&scope=")
	u.WriteString(string(scope))
	u.WriteString("&state=")
	u.WriteString(state)
	u.WriteString("#wechat_redirect")

	return u.String()
}

func (e *Engine) GetAuthInfo(authCode string) (*zjson.Res, error) {
	u := zstring.Buffer(3)
	u.WriteString(e.apiURL)

	appid := e.config.GetAppID()
	switch true {
	case e.IsWeapp():
		return (e.config.(*Weapp)).GetSessionKey(authCode, "authorization_code")
	case e.IsQy():
		u.WriteString("/cgi-bin/user/getuserinfo?access_token=")
		token, err := e.GetAccessToken()
		if err != nil {
			return nil, err
		}
		u.WriteString(token)
		u.WriteString("&code=")
		u.WriteString(authCode)
		json, err := httpResProcess(http.Post(u.String()))
		if err == nil {
			openid := json.Get("OpenId").String()
			j, err := zjson.Set(json.String(), "openid", openid)
			if err == nil {
				accessToken, _ := e.GetAccessToken()
				j, _ = zjson.Set(j, "access_token", accessToken)
				njson := zjson.Parse(j)
				return njson, nil
			}
		}
		return json, err
	case e.IsOpen():
		return nil, errors.New("not support")
	default:
		u.WriteString("/sns/oauth2/")
		u.WriteString("access_token?appid=")
		u.WriteString(appid)
		u.WriteString("&secret=")
		u.WriteString(e.config.GetSecret())
		u.WriteString("&code=")
		u.WriteString(authCode)
		u.WriteString("&grant_type=authorization_code")
	}

	return httpResProcess(http.Post(u.String()))

}
