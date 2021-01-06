package wechat

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	netHttp "net/http"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
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
		switch v[i].(type) {
		case map[string]string, SendData:
			v[i] = zhttp.BodyJSON(v[i])
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
		if err := xml.EscapeText(buf, zstring.String2Bytes(v)); err != nil {
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
	code := data.Get("errcode").String()
	if code == "0" || code == "" {
		return &data, nil
	}
	msg := data.Get("errmsg").String()
	return &zjson.Res{}, errors.New(code + ": " + msg)
}
