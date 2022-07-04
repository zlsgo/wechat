package wechat

import (
	"errors"
	"strings"

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
	b, err := httpProcess(r, e)
	if err != nil {
		return nil, err
	}
	return CheckResError(b)
}

func httpProcess(r *zhttp.Res, e error) ([]byte, error) {
	if e != nil {
		return nil, httpError{Code: -2, Msg: "网络请求失败"}
	}
	if r.StatusCode() != 200 {
		return nil, httpError{Code: -2, Msg: "接口请求失败: " + r.Response().Status}
	}
	return r.Bytes(), nil
}

func httpPayProcess(r *zhttp.Res, e error) (map[string]string, error) {
	b, err := httpProcess(r, e)
	if err != nil {
		return nil, err
	}

	x, err := ParseXML2Map(b)
	if err == nil {
		if code, ok := x["return_code"]; ok && code == "SUCCESS" {
			if resultCode, ok := x["result_code"]; ok && resultCode != "FAIL" {
				return x, nil
			}
		}
		msg, ok := x["err_code_des"]
		if !ok {
			if msg, ok = x["return_msg"]; !ok {
				msg = "未知错误"
			}
		}
		if strings.Contains(msg, "无效，请检查需要验收的case") {
			msg = "沙盒只支持指定金额, 如: 101 https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=23_13"
		}
		return map[string]string{}, errors.New(msg)
	}

	return x, err
}

func (e *Engine) checkTokenExpiration(err error) bool {
	if err != nil && ErrorCode(err) == 42001 {
		_, _ = e.cache.Delete(cacheToken)
		return true
	}
	return false
}
