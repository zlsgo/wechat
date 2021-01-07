package wechat

import (
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

func (e *Engine) HttpAccessTokenGet(url string, v ...interface{}) (j *zjson.Res, err error) {
	token, err := e.GetAccessToken()
	if err != nil {
		return nil, err
	}
	j, err = httpResProcess(http.Get(url, append(transformSendData(v), zhttp.QueryParam{"access_token": token})...))
	if e.checkTokenExpiration(err) {
		return e.HttpAccessTokenGet(url, v...)
	}
	return
}

func (e *Engine) HttpAccessTokenPost(url string, v ...interface{}) (j *zjson.Res, err error) {
	var token string
	token, err = e.GetAccessToken()
	if err != nil {
		return
	}
	j, err = httpResProcess(http.Post(url, append(transformSendData(v), zhttp.QueryParam{"access_token": token})...))
	if e.checkTokenExpiration(err) {
		return e.HttpAccessTokenPost(url, v...)
	}
	return
}

func httpResProcess(r *zhttp.Res, e error) (*zjson.Res, error) {
	if e != nil {
		return nil, httpError{Code: -2, Msg: "网络请求失败"}
	}
	if r.StatusCode() != 200 {
		return nil, httpError{Code: -2, Msg: "接口请求失败: " + r.Response().Status}
	}
	return CheckResError(r.Bytes())
}

func (e *Engine) checkTokenExpiration(err error) bool {
	if err != nil && ErrorCode(err) == 42001 {
		_, _ = e.cache.Delete(cacheToken)
		return true
	}
	return false
}
