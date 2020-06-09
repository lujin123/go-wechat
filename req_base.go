package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"go.uber.org/zap"
)

const (
	SignTypeMD5 = "MD5"
	TradeType   = "JSAPI"
)

type wxService struct {
	client Http
	key    string
	logger *zap.Logger
}

func (w wxService) SetLogger(log *zap.Logger) {
	w.logger = log
}

func (w wxService) RandString(n int) string {
	return RandStringBytesMaskImprSrc(n)
}

func (w wxService) Get(ctx context.Context, url string, f HandlerFunc) error {
	return w.DoReq(ctx, http.MethodGet, url, "", nil, f)
}

func (w wxService) PostJSON(ctx context.Context, url string, req interface{}, f HandlerFunc) (err error) {
	return w.DoReq(ctx, http.MethodPost, url, contentTypeJSON, req, f)
}

func (w wxService) PostXML(ctx context.Context, url string, req interface{}, f HandlerFunc) (err error) {
	return w.DoReq(ctx, http.MethodPost, url, contentTypeXML, req, f)
}

func (w wxService) DoReq(ctx context.Context, method, url string, contentType string, req interface{}, f HandlerFunc) (err error) {
	w.logger.Info("[wx] request", zap.String("url", url), zap.String("contentType", contentType), zap.Any("body", req))
	defer func() {
		if err != nil {
			w.logger.Error("[wx] request", zap.Error(err))
		}
	}()
	var (
		body    io.Reader
		headers map[string]string
	)
	switch contentType {
	case contentTypeXML:
		buf, err := xml.Marshal(&req)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(buf)
	case contentTypeJSON:
		buf, err := json.Marshal(&req)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(buf)
	default:
		if buf, ok := req.([]byte); ok {
			body = bytes.NewBuffer(buf)
		}
	}

	if contentType != "" {
		headers = map[string]string{
			"Content-Type": contentType,
		}
	}
	return w.client.Do(ctx, method, url, headers, body, f)
}

func (w wxService) sign(ctx context.Context, req interface{}) (string, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	var params map[string]string
	if err := json.Unmarshal(buf, &params); err != nil {
		return "", err
	}

	paramStr, err := GenParamStr(params)
	if err != nil {
		return "", err
	}
	stringSignTemp := paramStr + "&key=" + w.key
	return HashMd5(stringSignTemp), nil
}
