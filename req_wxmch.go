package wechat

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
)

const (
	mchPayUrl    = "https://api.mch.weixin.qq.com/mmpaymkttransfers/promotion/transfers"
	mchReqUrl    = "https://api.mch.weixin.qq.com/mmpaymkttransfers/gettransferinfo"
	mchRefundUrl = "https://api.mch.weixin.qq.com/secapi/pay/refund"
)

type MchService interface {
	ReqWxToMchPay(ctx context.Context, req *MchPayReq) (*MchPayResp, error)
	ReqMchPayment(ctx context.Context, tradeNo string) (*MchPaymentQueryResp, error)
	ReqPayRefund(ctx context.Context, req *MchPayRefundReq) (*MchPayRefundResp, error)
}

type (
	MchConfig struct {
		AppId       string
		MchId       string
		ApiKey      string
		CaCertFile  string
		ApiCertFile string
		ApiKeyFile  string
	}

	MchPayReq struct {
		XMLName        xml.Name `xml:"xml" json:"-"`
		MchAppID       string   `xml:"mch_appid" json:"mch_appid"`
		MchID          string   `xml:"mchid" json:"mchid"`
		NonceStr       string   `xml:"nonce_str" json:"nonce_str"`
		Sign           string   `xml:"sign" json:"sign"`
		PartnerTradeNO string   `xml:"partner_trade_no" json:"partner_trade_no"`
		OpenID         string   `xml:"openid" json:"openid"`
		CheckName      string   `xml:"check_name" json:"check_name"`
		Amount         int64    `xml:"amount" json:"amount,string"`
		Desc           string   `xml:"desc" json:"desc"`
		SpbillCreateIP string   `xml:"spbill_create_ip" json:"spbill_create_ip"`
	}

	MchPayResp struct {
		XMLName        xml.Name `xml:"xml"`
		ReturnCode     string   `xml:"return_code"`
		ReturnMsg      string   `xml:"return_msg"`
		MchAppID       string   `xml:"mch_appid"`
		MchID          string   `xml:"mchid"`
		NonceStr       string   `xml:"nonce_str"`
		ResultCode     string   `xml:"result_code"`
		ErrCode        string   `xml:"err_code"`
		ErrCodeDes     string   `xml:"err_code_des"`
		PartnerTradeNO string   `xml:"partner_trade_no"`
		PaymentNO      string   `xml:"payment_no"`
		PaymentTime    string   `xml:"payment_time"`
	}

	mchPaymentQueryReq struct {
		XMLName        xml.Name `xml:"xml" json:"-"`
		MchAppID       string   `xml:"appid" json:"appid"`
		MchID          string   `xml:"mch_id" json:"mch_id"`
		NonceStr       string   `xml:"nonce_str" json:"nonce_str"`
		Sign           string   `xml:"sign" json:"sign"`
		PartnerTradeNO string   `xml:"partner_trade_no" json:"partner_trade_no"`
	}

	MchPaymentQueryResp struct {
		XMLName        xml.Name `xml:"xml"`
		ReturnCode     string   `xml:"return_code"`
		ReturnMsg      string   `xml:"return_msg"`
		MchAppID       string   `xml:"appid"`
		MchID          string   `xml:"mch_id"`
		NonceStr       string   `xml:"nonce_str"`
		ResultCode     string   `xml:"result_code"`
		ErrCode        string   `xml:"err_code"`
		ErrCodeDes     string   `xml:"err_code_des"`
		PartnerTradeNO string   `xml:"partner_trade_no"`
		DetailID       string   `xml:"detail_id"`
		Status         string   `xml:"status"`
		Reason         string   `xml:"reason"`
		OpenID         string   `xml:"openid"`
		TransferName   string   `xml:"transferName"`
		PaymentAmount  int64    `xml:"payment_amount"`
		TransferTime   string   `xml:"transfer_time"`
		PaymentTime    string   `xml:"PaymentTime"`
		Desc           string   `xml:"desc"`
	}

	MchPayRefundReq struct {
		XMLName       xml.Name `xml:"xml" json:"-"`
		AppID         string   `xml:"appid" json:"appid"`
		MchID         string   `xml:"mch_id" json:"mch_id"`
		NonceStr      string   `xml:"nonce_str" json:"nonce_str"`
		Sign          string   `xml:"sign" json:"sign"`
		TransactionId string   `xml:"transaction_id" json:"transaction_id"`
		OutRefundNo   string   `xml:"out_refund_no" json:"out_refund_no"`
		TotalFee      int64    `xml:"total_fee" json:"total_fee,string"`
		RefundFee     int64    `xml:"refund_fee" json:"refund_fee,string"`
		RefundDesc    string   `xml:"refund_desc" json:"refund_desc"`
	}

	MchPayRefundResp struct {
		XMLName       xml.Name `xml:"xml"`
		ReturnCode    string   `xml:"return_code"`
		ReturnMsg     string   `xml:"return_msg"`
		MchAppID      string   `xml:"appid"`
		MchID         string   `xml:"mch_id"`
		NonceStr      string   `xml:"nonce_str"`
		ResultCode    string   `xml:"result_code"`
		ErrCode       string   `xml:"err_code"`
		ErrCodeDes    string   `xml:"err_code_des"`
		TransactionId string   `xml:"transaction_id"`
		OutTradeNo    string   `xml:"out_trade_no"`
		OutRefundNo   string   `xml:"out_refund_no"`
		RefundId      string   `xml:"refund_id"`
		RefundFee     int64    `xml:"refund_fee"`
		TotalFee      int64    `xml:"total_fee"`
		CashFee       int64    `xml:"cash_fee"`
	}

	wxMch struct {
		cfg *MchConfig
		wxService
	}
)

func NewWxMchService(cfg *MchConfig) *wxMch {
	s := &wxMch{
		cfg,
		wxService{
			client: nil,
			key:    cfg.ApiKey,
			logger: zapLogger,
		},
	}
	s.client = NewCtxHttpWithClient(s.TLSClient())
	zapLogger.Info("init wx mch service success...")
	return s
}

//企业付款到零钱接口
//接口文档：https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2
func (w wxMch) ReqWxToMchPay(ctx context.Context, req *MchPayReq) (*MchPayResp, error) {
	sign, err := w.sign(ctx, &req)
	if err != nil {
		return nil, err
	}
	req.Sign = sign

	var resp MchPayResp
	if err := w.PostXML(ctx, mchPayUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxmch] req wx to mch pay", zap.Any("body", resp))
	return &resp, nil
}

//企业付款到零钱查询接口
//接口文档：https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_3
func (w wxMch) ReqMchPayment(ctx context.Context, tradeNo string) (*MchPaymentQueryResp, error) {
	req := mchPaymentQueryReq{
		MchAppID:       w.cfg.AppId,
		MchID:          w.cfg.MchId,
		NonceStr:       w.RandString(32),
		Sign:           "",
		PartnerTradeNO: tradeNo,
	}
	sign, err := w.sign(ctx, &req)
	if err != nil {
		return nil, err
	}
	req.Sign = sign

	var resp MchPaymentQueryResp
	if err := w.PostXML(ctx, mchReqUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxmch] req mch payment", zap.Any("body", resp))
	return &resp, nil
}

//申请退款接口
//接口文档：https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=9_4
func (w wxMch) ReqPayRefund(ctx context.Context, req *MchPayRefundReq) (*MchPayRefundResp, error) {
	sign, err := w.sign(ctx, &req)
	if err != nil {
		return nil, err
	}
	req.Sign = sign

	var resp MchPayRefundResp
	if err := w.PostXML(ctx, mchRefundUrl, &req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return xml.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	w.logger.Info("[wxmch] req mch pay refund", zap.Any("body", resp))
	return &resp, nil
}

func (w wxMch) TLSClient() *http.Client {
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(w.cfg.CaCertFile)
	if err != nil {
		w.logger.Panic("[wx] read CACertFile", zap.Error(err))
	}
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair(w.cfg.ApiCertFile, w.cfg.ApiKeyFile)
	if err != nil {
		w.logger.Panic("[wx] LoadX509KeyPair", zap.Error(err))
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cliCrt},
		},
	}
	return &http.Client{
		Transport: tr,
	}
}
