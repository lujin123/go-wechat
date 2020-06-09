package wechat

import "testing"

var (
	payService PayService
)

func init() {
	cfg := PayConfig{
		AppId:     "wx3dfcc5eb4a8e335f",
		MchId:     "",
		ApiKey:    "",
		SignType:  "",
		TradeType: "",
	}
	payService = NewWxPay(&cfg, NewCtxHttp())
}

func TestWxPay_GenPrepay(t *testing.T) {

}
