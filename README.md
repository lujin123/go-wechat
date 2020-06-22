# 微信小程序sdk

用 `golang` 写的一个用于微信小程序支付以及小程序后端接口调用的`sdk`

## 接口

接口没有实现完整，目前实现的是一些用到比较多的接口

### 无需证书支付接口(`req_wxpay`)

- [x] 统一下单接口（`ReqUnifiedOrder`）
- [x] 订单查询接口（`ReqQueryOrder`）
- [x] 关闭订单接口（`ReqCloseOrder`）

### 需要证书支付接口(`req_wxmch`)

- [x] 企业付款到零钱接口（`ReqWxToMchPay`）
- [x] 企业付款到零钱查询接口（`ReqMchPayment`）
- [x] 申请退款接口（`ReqPayRefund`）

### 小程序接口(`req_wxmini`)

- [x] 获取`AccessToken`的接口（`ReqAccessToken`）
- [x] `code`换`session`接口（`ReqCode2Session`）
- [x] 发送订阅消息接口（`SendSubscribeMessage`）
- [x] 无限获取小程序码接口（`ReqWxCodeUnlimited`）
- [x] 校验图片是否含有违法违规内容接口（`CheckImage`）
- [x] 检查文本是否含有违法违规内容接口（`CheckMessage`）

### 工具方法

- [x] 生成小程序调用微信支付的预支付数据方法（`GenPrepay`）
- [x] 校验签名的方法（`VerifySign`）
- [x] 小程序即可设置token方法(`SetAccessToken`)

## 安装

```sh
> go get github.com/lujin123/go-wechat
```

## 使用

### 初始化

目前是有三个对象去分别处理微信小程序服务端接口、需要证书支付接口和无需证书支付接口

接口中使用的`http`服务是自己封装的一个带`context`的`http`，这样更加的灵活和方便

日志组件使用的是`uber`的`zap`库，这个库功能很强大，个人很喜欢，就直接用了。这样子会直接引入一个依赖，也许应该写一个接口去做，方便适配，
但是如果要将参数像`zap`那样序列化还挺麻烦，暂时就算了

#### 微信小程序
```go
package myapp

import (
    "github.com/lujin123/wechat"
)

func init() {
    cfg := wechat.MiniConfig{
        AppId:     "",
        AppSecret: "",
        SignType:  "",
        TradeType: "",
    }
    miniService := wechat.NewWxMiniService(&cfg, wechat.NewCtxHttp())
    //调用需要的方法
}
```

#### 微信无需证书支付
```go
package myapp

import (
    "github.com/lujin123/wechat"
)

func init() {
    cfg := wechat.PayConfig{
        AppId:"",
        MchId:"",
        ApiKey:"",
        SignType:"",
        TradeType:"",
    }
    payService := wechat.NewWxPayService(&cfg, wechat.NewCtxHttp())
    //调用需要的方法
}
```

#### 微信需证书支付
```go
package myapp

import (
    "github.com/lujin123/wechat"
)

func init() { 
    //这个服务初始化的时候需要传入微信商户后天的证书，用于生成一个带证书的客户端
    cfg := wechat.MchConfig{
        AppId:"",
        MchId:"",
        ApiKey:"",
        CaCertFile:"",
        ApiCertFile:"",
        ApiKeyFile:"",
    }
    mchService := wechat.NewWxMchService(&cfg)
    //调用需要的方法
}
```

### 接口

具体接口参考测试用例

## 最后

这是项目中需要用到，所以归总了下，方便其他的项目调用，现在直接用这个做个服务给外面调用，
尤其是关于`token`的过期问题，服务需要保证`token`有效，其他的服务调用即可，否则每次都要关心这个接口调用是否因为`token`失效原因失败了，
再获取`token`再调用，而且还要做好锁的问题，否则多个线程调用会导致老的`token`又失效的情况，灰常麻烦，具体的问题[参考官网](https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/access-token/auth.getAccessToken.html)
