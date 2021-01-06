package wechat

import (
	"encoding/xml"
	"errors"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	Reply struct {
		openid string
	}
	CDATA struct {
		Value string `xml:",cdata"`
	}
	ReplySt struct {
		Content      string `xml:"Content"`
		CreateTime   uint64
		FromUserName string
		MsgId        int
		MsgType      string
		Event        string
		ToUserName   string

		MediaId string
		// image
		PicUrl string

		// voice
		Format      string
		Recognition string

		// video or shortvideo
		ThumbMediaId string

		// location
		LocationX string `xml:"Location_X"`
		LocationY string `xml:"Location_Y"`
		Longitude string `xml:"Longitude"`
		Latitude  string `xml:"Latitude"`

		Scale string
		Label string

		// link
		Title       string
		Description string
		Url         string

		// Qy
		AgentID    string `xml:"AgentID"`
		isEncrypt  bool
		receiverID string
		received   *ReceivedSt
	}
)

type (
	ReceivedSt struct {
		echostr        string
		data           *ReplySt
		encrypt        string
		isEncrypt      bool
		signature      string
		timestamp      string
		nonce          string
		bodyData       []byte
		token          string
		encodingAesKey string
		msgSignature   string
	}
)

func (e *Engine) Reply(querys map[string]string,
	bodyData []byte) (received *ReceivedSt,
	err error) {
	received = &ReceivedSt{}
	received.echostr, _ = querys["echostr"]
	received.signature, _ = querys["signature"]
	received.timestamp, _ = querys["timestamp"]
	received.nonce, _ = querys["nonce"]
	received.encrypt, _ = querys["encrypt"]
	received.msgSignature, _ = querys["msg_signature"]
	received.token = e.GetToken()
	received.encodingAesKey = e.GetEncodingAesKey()
	received.isEncrypt = received.msgSignature != ""
	received.bodyData = bodyData
	return
}

func (r *ReceivedSt) Valid() (validMsg string, err error) {
	if r.isEncrypt {
		if r.msgSignature != sha1Signature(r.token, r.timestamp, r.nonce, r.echostr) {
			err = errors.New("decryption failed")
			return
		}
		var plaintext []byte
		plaintext, err = aesDecrypt(r.echostr, r.encodingAesKey)
		if err != nil {
			return
		}
		_, _, plaintext, _, err = parsePlainText(plaintext)
		if err != nil {
			return
		}
		validMsg = zstring.Bytes2String(plaintext)
	} else {
		if r.signature != sha1Signature(r.token, r.timestamp, r.nonce) {
			err = errors.New("decryption failed")
			return
		}
		validMsg = r.echostr
	}
	return
}

func (r *ReceivedSt) Data() (data *ReplySt, err error) {
	if r.data != nil {
		return r.data, nil
	}
	if r.isEncrypt {
		var arr XMLData
		arr, err = ParseXML2Map(r.bodyData)
		if err != nil {
			return
		}
		var plaintext []byte
		plaintext, err = aesDecrypt(arr["Encrypt"], r.encodingAesKey)
		if err != nil {
			return
		}
		var receiverID []byte
		_, _, plaintext, receiverID, err = parsePlainText(plaintext)
		if err != nil {
			return
		}

		log.Debug(zstring.Bytes2String(plaintext))
		err = xml.Unmarshal(plaintext, &data)
		if err == nil {
			data.received = r
			data.isEncrypt = true
			data.receiverID = zstring.Bytes2String(receiverID)
		}
	} else {
		// log.Debug(zstring.Bytes2String(r.bodyData))
		err = xml.Unmarshal(r.bodyData, &data)
	}
	return
}

func (t *ReplySt) ReplyCustom(fn func(r *ReplySt) (xml string)) string {
	reply := t.encrypt(fn(t))
	return reply
}

func (t *ReplySt) encrypt(content string) string {
	var err error
	if t.isEncrypt {
		data := map[string]string{}
		var encrypt []byte
		encrypt, err = aesEncrypt(MarshalPlainText(content, t.receiverID,
			zstring.Rand(16)),
			t.received.encodingAesKey)
		if err != nil {
			return ""
		}
		signature := sha1Signature(t.received.token, zstring.Bytes2String(encrypt), t.received.timestamp, t.received.nonce)
		data["Encrypt"] = zstring.Bytes2String(encrypt)
		data["MsgSignature"] = signature
		data["TimeStamp"] = t.received.timestamp
		data["Nonce"] = t.received.nonce
		reply, _ := FormatMap2XML(data)
		return reply
	}
	return content
}

func (t *ReplySt) ReplyText(content ...string) (reply string) {
	if len(content) == 0 {
		return "success"
	}
	data := map[string]string{
		"Content":      content[0],
		"CreateTime":   strconv.FormatInt(time.Now().Unix(), 10),
		"ToUserName":   t.FromUserName,
		"FromUserName": t.ToUserName,
		"MsgType":      "text",
	}
	reply, _ = FormatMap2XML(data)
	reply = t.encrypt(reply)
	return
}
