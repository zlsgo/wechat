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
