package wechat

import (
	"context"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	unifiedOrderUrl = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	closeOrderUrl   = "https://api.mch.weixin.qq.com/pay/closeorder"
	queryOrderUrl   = "https://api.mch.weixin.qq.com/pay/orderquery"
)

type PayService interface {
	// req function
	ReqUnifiedOrder(ctx context.Context, req *UnifiedOrderReq) (*UnifiedOrderResp, error)
	ReqQueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResp, error)
	ReqCloseOrder(ctx context.Context, tradeNo string) (*CloseOrderResp, error)

	// utils function
	GenPrepay(ctx context.Context, prepayId, nonceStr string) (*PrepayReturn, error)
	VerifySign(ctx context.Context, req *NotifyReq) bool
}

type (
	PayConfig struct {
		AppId     string
		MchId     string
		ApiKey    string
		SignType  string
		TradeType string
	}

	UnifiedOrderReq struct {
		XMLName        xml.Name `json:"-" xml:"xml"`
		AppId          string   `json:"appid" xml:"appid"`                       //小程序ID
		MchId          string   `json:"mch_id" xml:"mch_id"`                     //商户号
		DeviceInfo     string   `json:"device_info" xml:"device_info"`           //设备号
		NonceStr       string   `json:"nonce_str" xml:"nonce_str"`               //随机字符串
		Sign           string   `json:"sign" xml:"sign"`                         //签名
		SignType       string   `json:"sign_type" xml:"sign_type"`               //签名类型，默认为MD5，支持HMAC-SHA256和MD5
		Body           string   `json:"body" xml:"body"`                         //商品描述
		Detail         string   `json:"detail" xml:"detail"`                     //商品详情
		Attach         string   `json:"attach" xml:"attach"`                     //附加数据
		OutTradeNo     string   `json:"out_trade_no" xml:"out_trade_no"`         //商户订单号
		FeeType        string   `json:"fee_type" xml:"fee_type"`                 //标价币种
		TotalFee       int64    `json:"total_fee,string" xml:"total_fee"`        //标价金额
		SpbillCreateIp string   `json:"spbill_create_ip" xml:"spbill_create_ip"` //终端IP
		TimeStart      string   `json:"time_start" xml:"time_start"`             //交易起始时间
		TimeExpire     string   `json:"time_expire" xml:"time_expire"`           //交易结束时间
		GoodsTag       string   `json:"goods_tag" xml:"goods_tag"`               //订单优惠标记
		NotifyUrl      string   `json:"notify_url" xml:"notify_url"`             //通知地址
		TradeType      string   `json:"trade_type" xml:"trade_type"`             //交易类型
		OpenId         string   `json:"openid" xml:"openid"`                     //用户标识,trade_type=JSAPI，此参数必传，用户在商户appid下的唯一标识
	}

	UnifiedOrderResp struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
		AppID      string   `xml:"appid"`
		MchID      string   `xml:"mch_id"`
		NonceStr   string   `xml:"nonce_str"`
		Sign       string   `xml:"sign"`
		ResultCode string   `xml:"result_code"`
		ErrCode    string   `xml:"err_code"`
		ErrCodeDes string   `xml:"err_code_des"`
		TradeType  string   `xml:"trade_type"`
		PrepayId   string   `xml:"prepay_id"`
		CodeUrl    string   `xml:"code_url"`
	}

	CloseOrderReq struct {
		XMLName    xml.Name `json:"-" xml:"xml"`
		AppId      string   `json:"appid" xml:"appid"`
		MchId      string   `json:"mch_id" xml:"mch_id"`
		NonceStr   string   `json:"nonce_str" xml:"nonce_str"`
		OutTradeNo string   `json:"out_trade_no" xml:"out_trade_no"`
		Sign       string   `json:"sign" xml:"sign"`
		SignType   string   `json:"sign_type" xml:"sign_type"`
	}

	CloseOrderResp struct {
		ReturnCode string `xml:"return_code"`
		ReturnMsg  string `xml:"return_msg"`
		AppID      string `xml:"appid"`
		MchID      string `xml:"mch_id"`
		NonceStr   string `xml:"nonce_str"`
		Sign       string `xml:"sign"`
		ResultCode string `xml:"result_code"`
		ResultMsg  string `xml:"result_msg"`
		ErrCode    string `xml:"err_code"`
		ErrCodeDes string `xml:"err_code_des"`
	}

	PrepayReturn struct {
		AppId     string `json:"appId"`
		TimeStamp string `json:"timeStamp"`
		NonceStr  string `json:"nonceStr"`
		Package   string `json:"package"`
		SignType  string `json:"signType"`
		PaySign   string `json:"paySign"`
	}

	NotifyReq struct {
		XMLName       xml.Name `xml:"xml" json:"-"`
		ReturnCode    string   `xml:"return_code" json:"return_code"`
		ReturnMsg     string   `xml:"return_msg" json:"return_msg"`
		AppID         string   `xml:"appid" json:"appid"`
		MchID         string   `xml:"mch_id" json:"mch_id"`
		DeviceInfo    string   `xml:"device_info" json:"device_info"`
		NonceStr      string   `xml:"nonce_str" json:"nonce_str"`
		Sign          string   `xml:"sign" json:"sign"`
		SignType      string   `xml:"sign_type" json:"sign_type"`
		ResultCode    string   `xml:"result_code" json:"result_code"`
		ErrCode       string   `xml:"err_code" json:"err_code"`
		ErrCodeDes    string   `xml:"err_code_des" json:"err_code_des"`
		OpenId        string   `xml:"openid" json:"openid"`
		IsSubscribe   string   `xml:"is_subscribe" json:"is_subscribe"`
		TradeType     string   `xml:"trade_type" json:"trade_type"`
		BankType      string   `xml:"bank_type" json:"bank_type"`
		TotalFee      string   `xml:"total_fee" json:"total_fee"`
		FeeType       string   `xml:"fee_type" json:"fee_type"`
		CashFee       string   `xml:"cash_fee" json:"cash_fee"`
		CashFeeType   string   `xml:"cash_fee_type" json:"cash_fee_type"`
		TransactionId string   `xml:"transaction_id" json:"transaction_id"`
		OutTradeNo    string   `xml:"out_trade_no" json:"out_trade_no"`
		TimeEnd       string   `xml:"time_end" json:"time_end"`
	}

	NotifyResp struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
	}

	QueryOrderReq struct {
		XMLName    xml.Name `xml:"xml" json:"-"`
		AppID      string   `xml:"appid" json:"appid"`
		MchID      string   `xml:"mchid" json:"mchid"`
		OutTradeNo string   `xml:"out_trade_no" json:"out_trade_no"`
		NonceStr   string   `xml:"nonce_str" json:"nonce_str"`
		Sign       string   `xml:"sign" json:"sign"`
		SignType   string   `xml:"sign_type" json:"sign_type"`
	}

	QueryOrderResp struct {
		XMLName     xml.Name `xml:"xml" json:"-"`
		ReturnCode  string   `xml:"return_code" json:"return_code"`
		ReturnMsg   string   `xml:"return_msg" json:"return_msg"`
		AppID       string   `xml:"appid" json:"appid"`
		MchID       string   `xml:"mch_id" json:"mch_id"`
		NonceStr    string   `xml:"nonce_str" json:"nonce_str"`
		Sign        string   `xml:"sign" json:"sign"`
		ResultCode  string   `xml:"result_code" json:"result_code"`
		ErrCode     string   `xml:"err_code" json:"err_code"`
		ErrCodeDes  string   `xml:"err_code_des" json:"err_code_des"`
		DeviceInfo  string   `xml:"device_info" json:"device_info"`
		OpenId      string   `xml:"openid" json:"openid"`
		IsSubscribe string   `xml:"is_subscribe" json:"is_subscribe"`
		TradeType   string   `xml:"trade_type" json:"trade_type"`
		//SUCCESS—支付成功
		//REFUND—转入退款
		//NOTPAY—未支付
		//CLOSED—已关闭
		//REVOKED—已撤销（刷卡支付）
		//USERPAYING--用户支付中
		//PAYERROR--支付失败(其他原因，如银行返回失败)
		TradeState         string `xml:"trade_state" json:"trade_state"`           //交易状态
		TradeStateDesc     string `xml:"trade_state_desc" json:"trade_state_desc"` //交易状态描述
		BankType           string `xml:"bank_type" json:"bank_type"`
		TotalFee           int64  `xml:"total_fee" json:"total_fee"`
		SettlementTotalFee int64  `xml:"settlement_total_fee" json:"settlement_total_fee"`
		FeeType            string `xml:"fee_type" json:"fee_type"`
		CashFee            int64  `xml:"cash_fee" json:"cash_fee"`
		CashFeeType        string `xml:"cash_fee_type" json:"cash_fee_type"`
		TransactionId      string `xml:"transaction_id" json:"transaction_id"`
		OutTradeNo         string `xml:"out_trade_no" json:"out_trade_no"`
		TimeEnd            string `xml:"time_end" json:"time_end"`
	}
)

type wxPay struct {
	cfg *PayConfig
	wxService
}

func NewWxPayService(cfg *PayConfig, client Http) *wxPay {
	s := &wxPay{
		cfg,
		wxService{
			client: client,
			key:    cfg.ApiKey,
			logger: zapLogger,
		},
	}
	zapLogger.Info("init wx pay service success...")
	return s
}

// 统一下单接口
// 接口文档：https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=9_1
func (w wxPay) ReqUnifiedOrder(ctx context.Context, req *UnifiedOrderReq) (*UnifiedOrderResp, error) {
	var resp UnifiedOrderResp
	if err := w.PostXML(ctx, unifiedOrderUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxpay] unified order", zap.Any("resp", resp))
	return &resp, nil
}

// 订单查询接口
// 接口文档：https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=9_2
func (w wxPay) ReqQueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResp, error) {
	req := QueryOrderReq{
		AppID:      w.cfg.AppId,
		MchID:      w.cfg.MchId,
		OutTradeNo: tradeNo,
		NonceStr:   w.RandString(32),
		Sign:       "",
		SignType:   SignTypeMD5,
	}
	sign, err := w.sign(ctx, &req)
	if err != nil {
		return nil, err
	}
	req.Sign = sign
	var resp QueryOrderResp
	if err := w.PostXML(ctx, queryOrderUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxpay] query order", zap.Any("resp", resp))
	return &resp, nil
}

// 关闭订单
// 以下情况需要调用关单接口：商户订单支付失败需要生成新单号重新发起支付，要对原订单号调用关单，避免重复支付；系统下单后，用户支付超时，系统退出不再受理，避免用户继续，请调用关单接口。
// 接口文档：https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=9_3
func (w wxPay) ReqCloseOrder(ctx context.Context, tradeNo string) (*CloseOrderResp, error) {
	req := CloseOrderReq{
		AppId:      w.cfg.AppId,
		MchId:      w.cfg.MchId,
		NonceStr:   w.RandString(32),
		OutTradeNo: tradeNo,
		Sign:       "",
		SignType:   SignTypeMD5,
	}
	sign, err := w.sign(ctx, &req)
	if err != nil {
		return nil, err
	}
	req.Sign = sign

	var resp CloseOrderResp
	if err := w.PostXML(ctx, closeOrderUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxpay] close order", zap.Any("resp", resp))
	return &resp, nil
}

// 生成小程序预支付数据
func (w wxPay) GenPrepay(ctx context.Context, prepayId, nonceStr string) (*PrepayReturn, error) {
	if nonceStr == "" {
		nonceStr = w.RandString(32)
	}
	prepay := PrepayReturn{
		AppId:     w.cfg.AppId,
		TimeStamp: strconv.FormatInt(time.Now().Unix(), 10),
		NonceStr:  nonceStr,
		Package:   "prepay_id=" + prepayId,
		SignType:  SignTypeMD5,
		PaySign:   "",
	}
	sign, err := w.sign(ctx, &prepay)
	if err != nil {
		return nil, err
	}
	prepay.PaySign = sign
	return &prepay, nil
}

// 校验签名
func (w wxPay) VerifySign(ctx context.Context, req *NotifyReq) bool {
	oldSign := req.Sign
	req.Sign = ""
	sign, err := w.sign(ctx, req)
	if err != nil {
		w.logger.Error("[wxpay] verify sign", zap.Error(err))
		return false
	}
	return oldSign == sign
}
