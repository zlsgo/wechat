package wechat

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	Cfg interface {
		GetAppID() string
		GetToken() string
		GetEncodingAesKey() string
		getEngine() *Engine
		setEngine(*Engine)
		GetSecret() string
		getAccessToken() (data []byte, err error)
		getJsapiTicket() (data *zhttp.Res, err error)
	}
	Engine struct {
		config Cfg
		cache  *zcache.Table
		action string
	}
)

const (
	apiurl                     = "https://api.weixin.qq.com"
	qyurl                      = "https://qyapi.weixin.qq.com"
	cachePrefix                = "go_wechat_"
	cacheToken                 = "Token"
	cacheJsapiTicket           = "JsapiTicket"
	cacheComponentVerifyTicket = "componentVerifyTicket"
)

var (
	log   = zlog.New("[Wx] ")
	apps  = map[string]string{}
	debug bool
)

func init() {
	log.ResetFlags(zlog.BitLevel | zlog.BitTime)
	Debug(false)
}

func Debug(disable ...bool) {
	state := true
	if len(disable) > 0 {
		state = disable[0]
	}
	if state {
		debug = true
		log.SetLogLevel(zlog.LogDump)
	} else {
		debug = false
		log.SetLogLevel(zlog.LogWarn)
	}

}

var cacheData []byte

func LoadCacheData(path string) (err error) {
	var f os.FileInfo
	path = zfile.RealPath(path)
	f, err = os.Stat(path)
	if err != nil || f.IsDir() {
		return errors.New("file does not exist")
	}
	var data []byte
	var now = time.Now().Unix()
	data, _ = ioutil.ReadFile(path)
	cacheData = data
	zjson.ParseBytes(data).ForEach(func(key, value zjson.Res) bool {
		k := strings.Split(key.String(), "|")
		if len(k) < 2 || (k[0] == "" || k[1] == "") {
			return true
		}
		cacheName := cachePrefix + k[1] + k[0]
		cache := zcache.New(cacheName)
		apps[k[0]] = k[1]
		value.ForEach(func(key, value zjson.Res) bool {
			cachekey := key.String()
			log.Debug("载入缓存", cacheName, cachekey)
			switch cachekey {
			default:
				var t uint = 0
				lifespan := isSetCache(value, now)
				if lifespan == 0 {
					return true
				}
				t = uint(lifespan)
				cache.Set(cachekey, value.Get("content").String(), t)
			}
			return true
		})
		return true
	})
	return nil
}

func isSetCache(value zjson.Res, now int64) (diffTime int) {
	saveTime := value.Get("SaveTime").Int()
	outTime := value.Get("OutTime").Int()
	diffTime = outTime - (int(now) - saveTime)
	return
}

func SaveCacheData(path string) (json string, err error) {
	var file *os.File
	json = "{}"
	path = zfile.RealPath(path)
	if zfile.FileExist(path) {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		content, err := ioutil.ReadAll(file)
		if err == nil && zjson.ValidBytes(content) {
			json = zstring.Bytes2String(content)
		}
	} else {
		file, err = os.Create(path)
	}
	if err != nil {
		return
	}
	defer file.Close()
	now := time.Now().Unix()
	for k, v := range apps {
		log.Debug("SaveCacheData: ", cachePrefix+v+k)
		cache := zcache.New(cachePrefix + v + k)
		cache.ForEachRaw(func(key string, value *zcache.Item) bool {
			title := k + "\\|" + v
			log.Debug(title, key)
			if str, ok := value.Data().(string); ok {
				json, _ = zjson.Set(json, title+"."+key+".content", str)
			} else {
				json, _ = zjson.Set(json, title+"."+key, value.Data())
			}
			json, _ = zjson.Set(json, title+"."+key+".SaveTime", now)
			json, _ = zjson.Set(json, title+"."+key+".OutTime",
				value.RemainingLife().Seconds())

			return true
		})
	}
	saveData := zstring.String2Bytes(json)
	if len(saveData) == 2 && len(cacheData) > 2 {
		return
	}
	_, err = file.Write(saveData)

	return
}

func New(c Cfg) *Engine {
	action := ""
	appid := c.GetAppID()
	switch c.(type) {
	case *Open:
		action = "open"
	case *Qy:
		action = "qy"
	case *Mp:
		action = "mp"
	}
	engine := &Engine{
		cache:  zcache.New(cachePrefix + action + appid),
		config: c,
		action: action,
	}
	c.setEngine(engine)
	apps[appid] = action
	return engine
}

func (e *Engine) GetAction() Cfg {
	return e.config
}

func (e *Engine) GetAppId() string {
	return e.config.GetAppID()
}

func (e *Engine) GetSecret() string {
	return e.config.GetSecret()
}

func (e *Engine) IsMp() bool {
	return e.action == "mp"
}

func (e *Engine) IsQy() bool {
	return e.action == "qy"
}

func (e *Engine) IsOpen() bool {
	return e.action == "open"
}

func (e *Engine) GetToken() string {
	return e.config.GetToken()
}
func (e *Engine) GetEncodingAesKey() string {
	return e.config.GetEncodingAesKey()
}
