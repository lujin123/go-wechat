package wechat

import (
	"context"
	"io"
	"log"
	"net/http"
)

const (
	contentTypeXML  = "application/xml"
	contentTypeJSON = "application/json"
)

type HandlerFunc = func(response *http.Response, err error) error

type Http interface {
	Get(ctx context.Context, url string, f HandlerFunc) error
	Post(ctx context.Context, url, contentType string, body io.Reader, f HandlerFunc) error
	PostJSON(ctx context.Context, url string, body io.Reader, f HandlerFunc) error
	PostXML(ctx context.Context, url string, body io.Reader, f HandlerFunc) error
	Do(ctx context.Context, method, url string, headers map[string]string, body io.Reader, f HandlerFunc) error
}

type ctxHttp struct {
	client *http.Client
}

func NewCtxHttp() *ctxHttp {
	return &ctxHttp{
		client: http.DefaultClient,
	}
}

func NewCtxHttpWithClient(client *http.Client) *ctxHttp {
	if client == nil {
		client = http.DefaultClient
	}
	return &ctxHttp{
		client: client,
	}
}

func (h *ctxHttp) Get(ctx context.Context, url string, f HandlerFunc) error {
	return h.Do(ctx, http.MethodGet, url, nil, nil, f)
}

func (h *ctxHttp) Post(ctx context.Context, url, contentType string, body io.Reader, f HandlerFunc) error {
	header := map[string]string{
		"Content-Type": contentType,
	}
	return h.Do(ctx, http.MethodPost, url, header, body, f)
}
func (h *ctxHttp) PostJSON(ctx context.Context, url string, body io.Reader, f HandlerFunc) error {
	header := map[string]string{
		"Content-Type": contentTypeJSON,
	}
	return h.Do(ctx, http.MethodPost, url, header, body, f)
}
func (h *ctxHttp) PostXML(ctx context.Context, url string, body io.Reader, f HandlerFunc) error {
	header := map[string]string{
		"Content-Type": contentTypeXML,
	}
	return h.Do(ctx, http.MethodPost, url, header, body, f)
}

func (h *ctxHttp) Do(ctx context.Context, method, url string, headers map[string]string, body io.Reader, f HandlerFunc) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	return h.do(ctx, req, f)
}

func (h *ctxHttp) do(ctx context.Context, req *http.Request, f HandlerFunc) error {
	c := make(chan error)
	req = req.WithContext(ctx)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ctx http goroutine panic, error=%v", err)
			}
		}()
		select {
		case <-ctx.Done():
			log.Println("ctx http goroutine quit...")
			c <- ctx.Err()
			return
		default:
			c <- f(h.client.Do(req))
		}
	}()
	return <-c
}
