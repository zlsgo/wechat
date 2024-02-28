package main

import (
	"github.com/zlsgo/wechat"
)

// Wx 微信实例
var (
	Wx     *wechat.Engine
	WxOpen *wechat.Engine
	WxQy   *wechat.Engine
	Weapp  *wechat.Engine
	Pay    *wechat.Pay
)

func main() {
	// 开启调试日志
	wechat.Debug()

	// 加载文件缓存数据
	_ = wechat.LoadCacheData("wechat.json")

	// 支持公众号 企业微信 开放平台 小程序 微信支付
	Wx = wechat.New(&wechat.Mp{
		AppID:     "",
		AppSecret: "",
		Token:     "",
	})
	WxOpen = wechat.New(&wechat.Open{
		AppID:          "",
		AppSecret:      "",
		EncodingAesKey: "",
	})
	WxQy = wechat.New(&wechat.Qy{
		CorpID:         "",
		Secret:         "",
		Token:          "",
		EncodingAesKey: "",
	})
	Weapp = wechat.New(&wechat.Weapp{
		AppID:     "",
		AppSecret: "",
	})
	Pay = wechat.NewPay(wechat.Pay{
		MchId:    "",
		Key:      "",
		CertPath: "",
		KeyPath:  "",
	})
}

func SaveWxCacheData() (string, error) {
	// 保存缓存数据至文件
	return wechat.SaveCacheData("wechat.json")
}
