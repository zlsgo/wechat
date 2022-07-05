package wechat

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	netHttp "net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	// SendData Send Data
	SendData map[string]string
	// XMLData XML Data
	XMLData map[string]string
	request struct {
		request *netHttp.Request
		rawData []byte
	}
)

func transformSendData(v []interface{}) []interface{} {
	for i := range v {
		switch val := v[i].(type) {
		case string:
			v[i] = val
		case map[string]string, SendData, map[string]interface{}:
			v[i] = zhttp.BodyJSON(val)
		}
	}
	return v
}

func FormatMap2XML(m XMLData) (string, error) {
	buf := zutil.GetBuff()
	defer zutil.PutBuff(buf)
	if _, err := io.WriteString(buf, "<xml>"); err != nil {
		return "", err
	}
	for k, v := range m {
		if _, err := io.WriteString(buf, fmt.Sprintf("<%s>", k)); err != nil {
			return "", err
		}
		val := ztype.ToString(v)
		if err := xml.EscapeText(buf, zstring.String2Bytes(val)); err != nil {
			return "", err
		}
		if _, err := io.WriteString(buf, fmt.Sprintf("</%s>", k)); err != nil {
			return "", err
		}
	}
	if _, err := io.WriteString(buf, "</xml>"); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseXML2Map parse xml to map
func ParseXML2Map(b []byte) (m XMLData, err error) {
	var (
		d     = xml.NewDecoder(bytes.NewReader(b))
		depth = 0
		tk    xml.Token
		key   string
		buf   bytes.Buffer
	)
	m = XMLData{}
	for {
		tk, err = d.Token()
		if err != nil {
			if err == io.EOF {
				err = nil
				return
			}
			return
		}
		switch v := tk.(type) {
		case xml.StartElement:
			depth++
			switch depth {
			case 2:
				key = v.Name.Local
				buf.Reset()
			case 3:
				if err = d.Skip(); err != nil {
					return
				}
				depth--
				key = "" // key == "" indicates that the node with depth==2 has children
			}
		case xml.CharData:
			if depth == 2 && key != "" {
				buf.Write(v)
			}
		case xml.EndElement:
			if depth == 2 && key != "" {
				m[key] = buf.String()
			}
			depth--
		}
	}
}

// CheckResError CheckResError
func CheckResError(v []byte) (*zjson.Res, error) {
	data := zjson.ParseBytes(v)
	code := data.Get("errcode").Int()
	if code != 0 {
		errmsg := data.Get("errmsg").String()
		if errmsg == "" {
			return &zjson.Res{}, httpError{Code: code, Msg: "errcode: " + strconv.Itoa(code)}
		}
		return &zjson.Res{}, httpError{Code: code, Msg: errmsg}
	}
	return &data, nil
}

func paramFilter(uri string) string {
	if u, err := url.Parse(uri); err == nil {
		querys := u.Query()
		for k := range querys {
			if k == "code" || k == "state" || k == "scope" {
				delete(querys, k)
			}
		}
		u.RawQuery = querys.Encode()
		uri = u.String()
	}
	return uri
}

func sortParam(v map[string]interface{}, key string) string {
	l := len(v)
	keys := make([]string, 0, l)
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := zstring.Buffer(l * 3)
	for i := range keys {
		k := keys[i]
		s := ztype.ToString(v[k])
		if len(s) == 0 {
			continue
		}
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(s)
	}
	return b.String() + "&key=" + key
}

func signParam(v string, signType, key string) string {
	switch strings.ToUpper(signType) {
	case "SHA1":
		b := sha1.Sum(zstring.String2Bytes(v))
		return hex.EncodeToString(b[:])
	default:
		// MD5
		return strings.ToUpper(zstring.Md5(v))
	}
}
