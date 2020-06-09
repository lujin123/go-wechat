package wechat

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	token       = ""
	miniService MiniService
)

func init() {
	cfg := MiniConfig{
		AppId:     "",
		AppSecret: "",
		SignType:  "",
		TradeType: "",
	}
	miniService = NewWxMini(&cfg, NewCtxHttp())
}

func TestWxMini_ReqAccessToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := miniService.ReqAccessToken(ctx)
	assert.Nil(t, err)
	t.Log(resp)
	assert.Equal(t, "", resp.ErrMsg)
	miniService.SetAccessToken(resp.AccessToken)
}

func TestWxMini_ReqWxCodeUnlimited(t *testing.T) {
	miniService.SetAccessToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	codeBuff, err := miniService.ReqWxCodeUnlimited(ctx, &WxCodeUnlimitedReq{
		Scene: "ljabcjkabc",
	})
	assert.Nil(t, err)
	t.Log(codeBuff)
}

func TestWxMini_CheckMessage(t *testing.T) {
	miniService.SetAccessToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		Msg    string
		HasErr bool
	}{
		{"hello", false},
		{"习近平", true},
		{"任志强", false},
		{"郝海东", false},
	}
	for _, test := range tests {
		resp, err := miniService.CheckMessage(ctx, test.Msg)
		assert.Nil(t, err)
		if test.HasErr {
			assert.Equal(t, resp.ErrCode != 0, true, "error msg = %s", resp.ErrMsg)
		} else {
			assert.Equal(t, resp.ErrCode, 0)
		}
	}
}

func TestWxMini_ReqCode2Session(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := miniService.ReqCode2Session(ctx, "12343")
	assert.Nil(t, err)
	t.Logf("ReqCode2Session resp: %+v", resp)
}
