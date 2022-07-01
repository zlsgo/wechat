package wechat

import (
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

// GetAuthUserInfo 获取用户信息
// 企业微信需要使用 user_ticket 代替 openid
func (e *Engine) GetAuthUserInfo(openid, authAccessToken string) (json *zjson.Res, err error) {
	u := zstring.Buffer(6)
	u.WriteString(e.apiURL)
	switch true {
	case e.IsQy():
		u.WriteString("/cgi-bin/user/getuserdetail?access_token=")
		u.WriteString(authAccessToken)
		return httpResProcess(http.Post(u.String(), zhttp.BodyJSON(map[string]interface{}{"user_ticket": openid})))
	default:
		u.WriteString("/sns/userinfo?access_token=")
		u.WriteString(authAccessToken)
		u.WriteString("&openid=")
		u.WriteString(openid)
		return httpResProcess(http.Get(u.String()))
	}

}
