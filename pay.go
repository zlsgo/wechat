package wechat

import (
	"errors"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type Pay struct {
	MchId      string // 商户ID
	Key        string // 密钥
	CertPath   string // 证书路径
	KeyPath    string // 证书路径
	sandbox    bool   // 开启支付沙盒
	sandboxKey string
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

func (p PayOrder) build() map[string]interface{} {
	m := ztype.ToMapString(p)
	m["openid"] = p.openid
	m["total_fee"] = p.totalFee
	m["spbill_create_ip"] = p.spbillCreateIp

	return m
}

func NewPayOrder(openid string, totalFee uint, ip string, opts ...PayOrderOption) PayOrder {
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
		spbillCreateIp: ip,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

// NewPay 创建支付
func NewPay(p Pay) *Pay {
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

	data["sign"] = signParam(sortParam(data, p.Key), "MD5")

	sMap := make(map[string]string, len(data))
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

	key, ok := xmlData["sandbox_signkey"]
	if !ok {
		return "", errors.New("获取沙盒 Key 失败")
	}
	return key, nil
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
func (p *Pay) Orderquery(o Order) (map[string]string, error) {
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

	data["sign"] = signParam(sortParam(data, p.getKey()), "MD5")
	url := "https://api.mch.weixin.qq.com/pay/orderquery"
	if p.sandbox {
		url = "https://api.mch.weixin.qq.com/sandboxnew/pay/orderquery"
	}

	sMap := make(map[string]string, len(data))
	for k, val := range data {
		sMap[k] = ztype.ToString(val)
	}

	xml, err := FormatMap2XML(sMap)
	if err != nil {
		return nil, err
	}

	return httpPayProcess(http.Post(url, xml))
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
	data["sign"] = signParam(sortParam(data, p.getKey()), "MD5")

	sMap := make(map[string]string, len(data))
	for k, val := range data {
		sMap[k] = ztype.ToString(val)
	}

	xml, err := FormatMap2XML(sMap)
	if err != nil {
		return "", err
	}

	xmlData, err := httpPayProcess(http.Post(url, xml))
	if err != nil {
		return "", err
	}
	return xmlData["prepay_id"], nil
}

// JsSign 微信页面支付签名
func (p *Pay) JsSign(appid, prepayID string) map[string]interface{} {
	data := map[string]interface{}{
		"signType":  "MD5",
		"timeStamp": time.Now().Unix(),
		"nonceStr":  zstring.Rand(16),
		"package":   "prepay_id=" + prepayID,
		"appId":     appid,
	}

	data["paySign"] = signParam(sortParam(data, p.Key), "MD5")
	return data
}

// Notify 支付通知
// HTTP应答状态码需返回200或204，必须为https地址
func (p *Pay) Notify(raw string) (map[string]string, error) {
	data, err := ParseXML2Map(zstring.String2Bytes(raw))
	if err != nil {
		return nil, err
	}

	if return_code, ok := data["return_code"]; ok && return_code == "SUCCESS" {
		signData := make(map[string]interface{}, len(data))
		resultSign := ""
		signType := ""
		for key := range data {
			if key == "sign" {
				resultSign = data[key]
				continue
			}
			if key == "sign_type" {
				signType = data[key]
			}
			signData[key] = data[key]
		}

		sign := signParam(sortParam(signData, p.getKey()), signType)
		if resultSign != sign {
			return nil, errors.New("非法支付结果通用通知")
		}
	}

	return data, nil
}
