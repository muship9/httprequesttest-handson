package example

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Api struct {
	token string
	httpClient *http.Client
}

func New(token string, opts ...Option) *Api  {
	api := &Api{
		token:      token,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

type Option func(*Api)

func OptionHTTPClient(c *http.Client) Option {
	return func(api *Api) {
		api.httpClient = c
	}
}

type ResponseBody struct {
	Text string `json:text`
}

func (api *Api) Get(ctx context.Context) (*ResponseBody, error){
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", api.token))

	resp, err := api.request(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var body ResponseBody
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

func (api *Api) request(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	respCh := make(chan *http.Response)
	errCh := make(chan error)

	go func() {
		resp, err := api.httpClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	select {
	case resp := <-respCh:
		return resp, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, errors.New("HTTP request cancelled")
	}
}