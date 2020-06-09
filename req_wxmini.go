package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	code2sessionUrl     = "https://api.weixin.qq.com/sns/jscode2session"
	accessTokenUrl      = "https://api.weixin.qq.com/cgi-bin/token"
	subscribeMessageUrl = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
	wxCodeUnlimitedUrl  = "https://api.weixin.qq.com/wxa/getwxacodeunlimit"
	checkImageUrl       = "https://api.weixin.qq.com/wxa/img_sec_check"
	checkMsgUrl         = "https://api.weixin.qq.com/wxa/msg_sec_check"
)

var (
	ErrTokenMissing = errors.New("token missing")
)

type MiniService interface {
	SetAccessToken(token string)
	ReqCode2Session(ctx context.Context, code string) (*SessionResp, error)
	ReqAccessToken(ctx context.Context) (*AccessTokenResp, error)
	SendSubscribeMessage(ctx context.Context, req *SubscribeMessageReq) (*ErrorResp, error)
	ReqWxCodeUnlimited(ctx context.Context, req *WxCodeUnlimitedReq) ([]byte, error)
	CheckImage(ctx context.Context, media []byte) (*ErrorResp, error)
	CheckMessage(ctx context.Context, msg string) (*ErrorResp, error)
}

type (
	MiniConfig struct {
		AppId     string
		AppSecret string
		SignType  string
		TradeType string
	}
	ErrorResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	SessionResp struct {
		ErrorResp
		OpenId     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionId    string `json:"unionid"`
	}
	AccessTokenResp struct {
		ErrorResp
		AccessToken string `json:"access_token"` //获取到的凭证
		ExpiresIn   int64  `json:"expires_in"`   //凭证有效时间，单位：秒。目前是7200秒之内的值。
	}
	SubscribeMessageReq struct {
		Touser           string                 `json:"touser"`
		TemplateId       string                 `json:"template_id"`
		Page             string                 `json:"page"`
		Data             map[string]interface{} `json:"data"`
		MiniprogramState string                 `json:"miniprogram_state"`
		Lang             string                 `json:"lang"`
	}

	WxCodeUnlimitedReq struct {
		Scene     string `json:"scene"`
		Page      string `json:"page"`
		Width     int    `json:"width"`
		AutoColor bool   `json:"auto_color"`
		LineColor struct {
			R int `json:"r"`
			G int `json:"g"`
			B int `json:"b"`
		} `json:"line_color"`
		IsHyaline bool `json:"is_hyaline"`
	}
)

type wxMini struct {
	cfg   *MiniConfig
	token string
	wxService
}

func NewWxMini(cfg *MiniConfig, client Http) *wxMini {
	s := &wxMini{
		cfg: cfg,
		wxService: wxService{
			client: client,
			logger: zapLogger,
		},
	}
	zapLogger.Info("init wx mini service success...")
	return s
}

//设置access_token
func (w *wxMini) SetAccessToken(token string) {
	w.token = token
}

// 登录凭证校验。通过 wx.login 接口获得临时登录凭证 code 后传到开发者服务器调用此接口完成登录流程
// 文档地址：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
func (w wxMini) ReqCode2Session(ctx context.Context, jsCode string) (*SessionResp, error) {
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", code2sessionUrl, w.cfg.AppId, w.cfg.AppSecret, jsCode)
	var sessionResp SessionResp
	if err := w.Get(ctx, url, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return json.NewDecoder(response.Body).Decode(&sessionResp)
	}); err != nil {
		return nil, err
	}

	return &sessionResp, nil
}

//获取小程序全局唯一后台接口调用凭据（access_token）
//接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html
func (w wxMini) ReqAccessToken(ctx context.Context) (*AccessTokenResp, error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", accessTokenUrl, w.cfg.AppId, w.cfg.AppSecret)
	var resp AccessTokenResp
	if err := w.Get(ctx, url, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		return json.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}

	return &resp, nil
}

//发送订阅消息
//接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.send.html
func (w wxMini) SendSubscribeMessage(ctx context.Context, req *SubscribeMessageReq) (*ErrorResp, error) {
	if err := w.checkToken(); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s?access_token=%s", subscribeMessageUrl, w.token)
	var resp ErrorResp
	if err := w.PostJSON(ctx, url, req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}

		return json.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	return &resp, nil
}

//获取小程序码，适用于需要的码数量极多的业务场景。通过该接口生成的小程序码，永久有效，数量暂无限制
//接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/qr-code/wxacode.getUnlimited.html
func (w wxMini) ReqWxCodeUnlimited(ctx context.Context, req *WxCodeUnlimitedReq) ([]byte, error) {
	if err := w.checkToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?access_token=%s", wxCodeUnlimitedUrl, w.token)
	var buff []byte
	if err := w.PostJSON(ctx, url, req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}
		buff, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		//检查是否为异常
		var resp ErrorResp
		//如果无法解析，就认为是二维码，能解析出来就是异常结果
		//撒币微信，数据结构搞成一样的会死？
		if err := json.Unmarshal(buff, &resp); err != nil {
			return nil
		}
		return errors.New(resp.ErrMsg)
	}); err != nil {
		return nil, err
	}
	return buff, nil
}

//校验一张图片是否含有违法违规内容。
//接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.imgSecCheck.html
func (w wxMini) CheckImage(ctx context.Context, media []byte) (*ErrorResp, error) {
	if err := w.checkToken(); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s?access_token=%s", checkImageUrl, w.token)

	var bodyBuff bytes.Buffer
	bodyWriter := multipart.NewWriter(&bodyBuff)
	if err := bodyWriter.WriteField("media", string(media)); err != nil {
		return nil, err
	}
	var resp ErrorResp
	if err := w.DoReq(ctx, http.MethodPost, url, bodyWriter.FormDataContentType(), bodyBuff.Bytes(), func(response *http.Response, err error) error {
		if err != nil {
			return err
		}

		return json.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}

	return &resp, nil
}

//检查一段文本是否含有违法违规内容。
//接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.msgSecCheck.html
func (w wxMini) CheckMessage(ctx context.Context, msg string) (*ErrorResp, error) {
	if err := w.checkToken(); err != nil {
		return nil, err
	}
	req := map[string]string{
		"content": msg,
	}
	url := fmt.Sprintf("%s?access_token=%s", checkMsgUrl, w.token)
	var resp ErrorResp
	if err := w.PostJSON(ctx, url, req, func(response *http.Response, err error) error {
		if err != nil {
			return err
		}

		return json.NewDecoder(response.Body).Decode(&resp)
	}); err != nil {
		return nil, err
	}
	return &resp, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////
func (w wxMini) checkToken() error {
	if w.token == "" {
		return ErrTokenMissing
	}
	return nil
}
