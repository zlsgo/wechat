package wechat

import (
	"errors"
	"strings"

	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/ztype"
)

type RouterOption struct {
	Prefix              string
	JsapiTicketCallback func(*znet.Context, ztype.Map, error)
}

func (w *Engine) Router(r *znet.Engine, opt RouterOption) {
	opt.Prefix = strings.TrimRight(opt.Prefix, "/")

	r.GET(opt.Prefix+"/js_ticket", func(c *znet.Context) {
		opt.JsapiTicketCallback(getJsapiTicket(w, c))
	})
}

func getJsapiTicket(wx *Engine, c *znet.Context) (*znet.Context, ztype.Map, error) {
	jsapiTicket, err := wx.GetJsapiTicket()
	if err != nil {
		return c, nil, errors.New(ErrorMsg(err))
	}
	url := c.Host(true)
	jsSign, err := wx.GetJsSign(url)
	if err != nil {
		return c, nil, errors.New(ErrorMsg(err))
	}

	return c, map[string]interface{}{
		"jsapiTicket": jsapiTicket,
		"jsSign":      jsSign,
		"url":         url,
	}, nil
}
