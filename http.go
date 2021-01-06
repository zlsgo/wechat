package wechat

import (
	"strconv"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
)

var http = zhttp.New()

func init() {
	http.DisableChunke()
}

func (e *Engine) Http() *zhttp.Engine {
	return http
}

func (e *Engine) HttpAccessTokenGet(url string, v ...interface{}) (*zjson.Res, error) {
	token, err := e.GetAccessToken()
	if err != nil {
		return nil, err
	}
	v = append(transformSendData(v), zhttp.QueryParam{"access_token": token})
	return httpResProcess(http.Get(url, v...))
}

func (e *Engine) HttpAccessTokenPost(url string, v ...interface{}) (*zjson.Res, error) {
	token, err := e.GetAccessToken()
	if err != nil {
		return nil, err
	}

	v = append(transformSendData(v), zhttp.QueryParam{"access_token": token})
	return httpResProcess(http.Post(url, v...))
}

func httpResProcess(r *zhttp.Res, e error) (*zjson.Res, error) {
	if e != nil {
		return nil, httpError{Code: -2, Msg: "网络请求失败"}
	}
	if r.StatusCode() != 200 {
		return nil, httpError{Code: -2, Msg: "接口请求失败: " + r.Response().Status}
	}
	json := zjson.ParseBytes(r.Bytes())
	errcode := json.Get("errcode").Int()
	if errcode != 0 {
		errmsg := json.Get("errmsg").String()
		if errmsg == "" {
			errmsg = "errcode: " + strconv.Itoa(errcode)
		}
		return &json, httpError{Code: errcode, Msg: errmsg}
	}

	return &json, nil
}
