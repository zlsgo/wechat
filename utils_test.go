package wechat

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestUtilsParamFlter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	for r, w := range map[string]string{
		"https://api.weixin.qq.com/?code=1&code=2&test=3": "https://api.weixin.qq.com/?test=3",
	} {
		tt.Equal(w, paramFilter(r))
	}
}

func TestUtilsKsortParam(t *testing.T) {
	tt := zlsgo.NewTest(t)
	for r, w := range map[string]map[string]interface{}{
		"a=1&b=dd&er=6&z=222&key=999": {
			"a":  1,
			"b":  "dd",
			"z":  222,
			"er": 6,
		},
	} {
		param := sortParam(w, "999")
		tt.Equal(r, param)
	}
}
