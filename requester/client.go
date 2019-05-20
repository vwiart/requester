package requester

import (
	"context"
	"io"
	"net/http"
)

type Client interface {
	Do(ctx context.Context, funcs ...func(*ClientParam)) (io.ReadCloser, error)
}

type client struct {
}

func New() Client {
	return client{}
}

type ClientParam struct {
	client        http.Client
	method        string
	url           string
	body          io.Reader
	headers       map[string]string
	statusHandler func(r *http.Response) error
}

func defaultStatusHandler(r *http.Response) error {
	return nil
}

type noopCloser struct{}

func (c noopCloser) Read([]byte) (int, error) { return 0, nil }

func (c noopCloser) Close() error { return nil }

func (c client) Do(ctx context.Context,
	funcs ...func(*ClientParam)) (io.ReadCloser, error) {
	params := ClientParam{
		statusHandler: defaultStatusHandler,
	}
	for _, apply := range funcs {
		apply(&params)
	}

	req, err := http.NewRequest(params.method, params.url, params.body)
	if err != nil {
		return noopCloser{}, err
	}

	for k, v := range params.headers {
		req.Header.Set(k, v)
	}
	req = req.WithContext(ctx)

	resp, err := params.client.Do(req)
	if err != nil {
		return noopCloser{}, err
	}

	if err := params.statusHandler(resp); err != nil {
		return noopCloser{}, err
	}

	return resp.Body, nil
}

func Method(method string) func(*ClientParam) {
	return func(params *ClientParam) {
		params.method = method
	}
}

func URL(url string) func(*ClientParam) {
	return func(params *ClientParam) {
		params.url = url
	}
}

func Body(body io.Reader) func(*ClientParam) {
	return func(params *ClientParam) {
		params.body = body
	}
}
