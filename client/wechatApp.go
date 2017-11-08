package client

import (

	"fmt"
	"github.com/qiuchengw/gopay/common"
	"github.com/qiuchengw/gopay/util"
	"strings"
	"time"
	"errors"
)

var defaultWechatAppClient *WechatAppClient

func InitWxAppClient(c *WechatAppClient) {
	defaultWechatAppClient = c
}

// DefaultWechatAppClient 默认微信app客户端
func DefaultWechatAppClient() *WechatAppClient {
	return defaultWechatAppClient
}

// WechatAppClient 微信app支付
type WechatAppClient struct {
	AppID       string // AppID
	MchID       string // 商户号ID
	CallbackURL string // 回调地址
	Key         string // 密钥
	PayURL      string // 支付地址
}

// Pay 支付
func (this *WechatAppClient) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = this.AppID
	m["mch_id"] = this.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = TruncatedText(charge.Describe,32)
	m["out_trade_no"] = charge.TradeNum
	m["total_fee"] = fmt.Sprintf("%d", int(charge.MoneyFee*100))
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "APP"
	m["sign_type"] = "MD5"

	sign ,err := WechatGenSign(this.Key,m)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.sign: " + err.Error())
	}

	m["sign"] = sign

	xmlRe ,err := PostWechat(this.PayURL,m)
	if  err != nil{
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["appid"] = this.AppID
	c["partnerid"] = this.MchID
	c["prepayid"] = xmlRe.PrepayID
	c["package"] = "Sign=WXPay"
	c["noncestr"] = util.RandomStr()
	c["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	sign2 ,err := WechatGenSign(this.Key,c)
	if err != nil {
		return map[string]string{}, errors.New("WechatApp.paySign: " + err.Error())
	}
	c["paySign"] = strings.ToUpper(sign2)

	return c, nil
}

// QueryOrder 查询订单
func (this *WechatAppClient) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
	var m = make(map[string]string)
	m["appid"] = this.AppID
	m["mch_id"] = this.MchID
	m["out_trade_no"] = tradeNum
	m["nonce_str"] = util.RandomStr()

	sign ,err := WechatGenSign(this.Key,m)
	if err != nil {
		return common.WeChatQueryResult{}, err
	}

	m["sign"] = sign

	return PostWechat("https://api.mch.weixin.qq.com/pay/orderquery",m)
}
