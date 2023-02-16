package wechat

import (
	"errors"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type Pay struct {
	MchId      string // 商户ID
	Key        string // V2密钥
	CertPath   string // 证书路径
	KeyPath    string // 证书路径
	prikey     string // 私钥内容
	sandbox    bool   // 开启支付沙盒
	sandboxKey string
	http       *zhttp.Engine
}

type OrderCondition struct {
	OutTradeNo    string `json:"out_trade_no,omitempty"`
	TransactionId string `json:"transaction_id,omitempty"`
}

type PayOrder struct {
	Appid          string `json:"appid,omitempty"`
	DeviceInfo     string `json:"device_info,omitempty"`
	NonceStr       string `json:"nonce_str,omitempty"`
	SignType       string `json:"sign_type,omitempty"`
	Body           string `json:"body,omitempty"`
	OutTradeNo     string `json:"out_trade_no,omitempty"`
	FeeType        string `json:"fee_type,omitempty"`
	totalFee       uint
	spbillCreateIp string
	TradeType      string `json:"trade_type,omitempty"`
	openid         string
}
type PayOrderOption func(*PayOrder)

func (p PayOrder) GetOutTradeNo() string {
	return p.OutTradeNo
}

func (p PayOrder) build() ztype.Map {
	m := ztype.ToMapString(p)
	m["openid"] = p.openid
	m["total_fee"] = p.totalFee
	m["spbill_create_ip"] = p.spbillCreateIp

	return m
}

// NewPayOrder 支付订单
func NewPayOrder(openid string, totalFee uint, ip string, body string, opts ...PayOrderOption) PayOrder {
	outTradeNo := zstring.Md5(openid)[0:12] + strconv.Itoa(int(time.Now().Unix())) + zstring.Rand(10)
	p := PayOrder{
		DeviceInfo:     "WEB",
		NonceStr:       zstring.Rand(16),
		SignType:       "MD5",
		FeeType:        "CNY",
		TradeType:      "JSAPI",
		openid:         openid,
		OutTradeNo:     outTradeNo,
		totalFee:       totalFee,
		Body:           body,
		spbillCreateIp: ip,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

type RefundOrder struct {
	Appid         string `json:"appid,omitempty"`
	DeviceInfo    string `json:"device_info,omitempty"`
	NonceStr      string `json:"nonce_str,omitempty"`
	SignType      string `json:"sign_type,omitempty"`
	OutRefundNo   string `json:"out_refund_no,omitempty"`
	RefundFeeType string `json:"refund_fee_type,omitempty"`
	totalFee      uint
	refundFee     uint
	OutTradeNo    string `json:"out_trade_no,omitempty"`
	TransactionId string `json:"transaction_id,omitempty"`
}

type RefundOrderOption func(*RefundOrder)

func (p RefundOrder) GetOutRefundNo() string {
	return p.OutRefundNo
}

func (p RefundOrder) build() map[string]interface{} {
	m := ztype.ToMapString(p)
	m["total_fee"] = p.totalFee
	m["refund_fee"] = p.refundFee

	return m
}

// NewRefundOrder 退款订单
func NewRefundOrder(totalFee, refundFee uint, condition OrderCondition, opts ...RefundOrderOption) RefundOrder {
	m := condition.OutTradeNo
	if len(m) == 0 {
		m = condition.TransactionId
	}
	OutRefundNo := zstring.Md5(m)[0:12] + strconv.Itoa(int(time.Now().Unix())) + zstring.Rand(10)
	p := RefundOrder{
		DeviceInfo:    "WEB",
		NonceStr:      zstring.Rand(16),
		SignType:      "MD5",
		RefundFeeType: "CNY",
		OutRefundNo:   OutRefundNo,
		OutTradeNo:    condition.OutTradeNo,
		TransactionId: condition.TransactionId,
		totalFee:      totalFee,
		refundFee:     refundFee,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

// NewPay 创建支付
func NewPay(p Pay) *Pay {
	p.http = zhttp.New()
	if len(p.CertPath) > 0 && len(p.KeyPath) > 0 {
		p.http.TlsCertificate(zhttp.Certificate{
			CertFile: p.CertPath,
			KeyFile:  p.KeyPath,
		})
	}
	return &p
}

func (p *Pay) Sandbox(enable bool) *Pay {
	p.sandbox = enable
	return p
}

func (p *Pay) GetSandboxSignkey() (string, error) {
	data := map[string]interface{}{
		"mch_id":    p.MchId,
		"key":       p.Key,
		"sign_type": "MD5",
		"nonce_str": zstring.Rand(16),
	}

	data["sign"] = signParam(sortParam(data, p.Key), "MD5", "")

	sMap := make(ztype.Map, len(data))
	for k, val := range data {
		sMap[k] = ztype.ToString(val)
	}

	xml, _ := FormatMap2XML(sMap)
	res, err := http.Post("https://api.mch.weixin.qq.com/sandboxnew/pay/getsignkey", xml)
	if err != nil {
		return "", err
	}

	xmlData, err := ParseXML2Map(res.Bytes())
	if err != nil {
		return "", err
	}

	key := xmlData.Get("sandbox_signkey").String()
	if key == "" {
		return "", errors.New("获取沙盒 Key 失败")
	}
	return key, nil
}

func (p *Pay) prikeyText() string {
	if len(p.prikey) == 0 {
		if prikey, err := zfile.ReadFile(p.KeyPath); err == nil {
			p.prikey = zstring.Bytes2String(prikey)
		}
	}
	return p.prikey
}

func (p *Pay) getKey() string {
	if !p.sandbox {
		return p.Key
	}
	if len(p.sandboxKey) == 0 {
		p.sandboxKey, _ = p.GetSandboxSignkey()
	}
	return p.sandboxKey
}

type Order struct {
	TransactionID string
	OutTradeNo    string
}

// Orderquery 订单查询
func (p *Pay) Orderquery(o Order) (ztype.Map, error) {
	if len(o.OutTradeNo) == 0 && len(o.TransactionID) == 0 {
		return nil, errors.New("out_trade_no、transaction_id 至少填一个")
	}

	data := map[string]interface{}{
		"mch_id":         p.MchId,
		"appid":          "wx591bf582cee71574",
		"nonce_str":      zstring.Rand(32),
		"sign_type":      "MD5",
		"transaction_id": o.TransactionID,
		"out_trade_no":   o.OutTradeNo,
	}

	data["sign"] = signParam(sortParam(data, p.getKey()), "MD5", "")
	url := "https://api.mch.weixin.qq.com/pay/orderquery"
	if p.sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	}

	sMap := make(ztype.Map, len(data))
	for k, val := range data {
		sMap[k] = ztype.ToString(val)
	}

	xml, err := FormatMap2XML(sMap)
	if err != nil {
		return nil, err
	}

	return httpPayProcess(p.http.Post(url, xml))
}

// UnifiedOrder 统一下单
func (p *Pay) UnifiedOrder(appid string, order PayOrder, notifyUrl string) (prepayID string, err error) {
	url := "https://api.mch.weixin.qq.com/pay/unifiedorder"
	if p.sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	}

	data := order.build()
	data["notify_url"] = notifyUrl
	data["mch_id"] = p.MchId
	data["appid"] = appid
	data["sign"] = signParam(sortParam(data, p.getKey()), "MD5", "")

	sMap := make(ztype.Map, len(data))
	for k := range data {
		sMap[k] = data.Get(k).String()
	}

	xml, err := FormatMap2XML(sMap)
	if err != nil {
		return "", err
	}

	xmlData, err := httpPayProcess(p.http.Post(url, xml))
	if err != nil {
		return "", err
	}
	return xmlData.Get("prepay_id").String(), nil
}

// Refund 申请退款
func (p *Pay) Refund(appid string, order RefundOrder, notifyUrl string) (refundID string, err error) {
	url := "https://api.mch.weixin.qq.com/secapi/pay/refund"
	if p.sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/refund"
	}

	data := order.build()
	data["notify_url"] = notifyUrl
	data["mch_id"] = p.MchId
	data["appid"] = appid
	data["sign"] = signParam(sortParam(data, p.getKey()), "MD5", "")

	sMap := make(ztype.Map, len(data))
	for k, val := range data {
		sMap[k] = ztype.ToString(val)
	}

	xml, err := FormatMap2XML(sMap)
	if err != nil {
		return "", err
	}

	xmlData, err := httpPayProcess(p.http.Post(url, xml))
	if err != nil {
		return "", err
	}

	return xmlData.Get("refund_id").String(), nil
}

// JsSign 微信页面支付签名
func (p *Pay) JsSign(appid, prepayID string) map[string]interface{} {
	data := map[string]interface{}{
		"signType":  "MD5",
		"timeStamp": strconv.Itoa(int(time.Now().Unix())),
		"nonceStr":  zstring.Rand(16),
		"package":   "prepay_id=" + prepayID,
		"appId":     appid,
	}

	key := p.prikeyText()
	data["paySign"] = signParam(sortParam(data, p.Key), "MD5", key)
	return data
}

type NotifyType uint

const (
	UnknownNotify NotifyType = iota
	PayNotify
	RefundNotify
)

type NotifyResult struct {
	Type     NotifyType
	Data     ztype.Map
	Response []byte
}

// Notify 支付/退款通知
func (p *Pay) Notify(raw string) (result *NotifyResult, err error) {
	result = &NotifyResult{
		Type: UnknownNotify,
	}

	defer func() {
		if err != nil {
			result.Response = []byte(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[` + err.Error() + `]]></return_msg></xml>`)
		} else {
			result.Response = []byte(`<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`)
		}
	}()

	var data ztype.Map
	data, err = ParseXML2Map(zstring.String2Bytes(raw))
	if err != nil {
		return
	}

	info := data.Get("req_info").String()
	if info != "" {
		result.Type = RefundNotify
		var plain []byte
		plain, err = aesECBDecrypt(info, zstring.Md5(p.getKey()))
		if err != nil {
			return
		}

		if d, err := ParseXML2Map(plain); err == nil {
			for k := range d {
				data[k] = d[k]
			}
			delete(data, "req_info")
		}
	} else {
		result.Type = PayNotify
	}

	result.Data = data

	if returnCode, ok := data["return_code"]; ok && returnCode == "SUCCESS" {
		signData := make(map[string]interface{}, len(data))
		resultSign := ""
		signType := ""
		for key := range data {
			if key == "sign" {
				resultSign = data.Get(key).String()
				continue
			}
			if key == "sign_type" {
				signType = data.Get(key).String()
			}
			signData[key] = data[key]
		}
		if len(signType) > 0 {
			sign := signParam(sortParam(signData, p.getKey()), signType, "")
			if resultSign != sign {
				err = errors.New("非法支付结果通用通知")
				return
			}
		}
	}

	return
}
