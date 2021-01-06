package wechat_test

import (
	"testing"

	"github.com/zlsgo/wechat"
)

func TestWechat(t *testing.T) {
	wx := wechat.New(&wechat.Mp{
		AppID:     "wx9d1fcb71007a71b0",
		AppSecret: "c4132441ded3301bda2d2373609959e1",
	})
	t.Log(wx.GetAccessToken())
}

func TestApi(t *testing.T) {

	wx := wechat.New(&wechat.Mp{
		AppID:     "wx9d1fcb71007a71b0",
		AppSecret: "c4132441ded3301bda2d2373609959e1",
	})

	res, err := wx.HttpAccessTokenPost("https://api.weixin.qq.com/cgi-bin/shorturl", map[string]string{
		"action":   "long2short",
		"long_url": "https://api.weixin.qq.com",
	})
	if err != nil {
		t.Fatal(wechat.ErrorMsg(err))
	}

	t.Log(res.String())

	res, err = wx.HttpAccessTokenPost("https://api.weixin.qq.com/cgi-bin/shorturl", map[string]string{
		"action": "long2short",
	})
	if err == nil {
		t.Fail()

	}
	t.Log(wechat.ErrorMsg(err))
	t.Log(res.String())
}
