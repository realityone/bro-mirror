package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/realityone/bro-mirror/internal/conf"
	"golang.org/x/exp/slices"
	"golang.org/x/net/http2"
)

type Upstream interface {
	GetClient() (*http.Client, string)
	Remote() string
}

type Mirror struct {
	client  *http.Client
	baseURL url.URL
	logger  log.Logger
}

func (m *Mirror) GetClient() (*http.Client, string) {
	return m.client, m.baseURL.String()
}
func (m *Mirror) Remote() string {
	return strings.TrimPrefix(m.baseURL.Host, "api.")
}

func makeH2Transport(cfg *conf.Mirror) *http2.Transport {
	transport := &http2.Transport{
		DisableCompression: true,
	}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	if cfg.WithoutTls {
		transport.AllowHTTP = true
		transport.DialTLSContext = func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, network, addr)
		}
	}
	return transport
}

func makeTransport(cfg *conf.Mirror) *http.Transport {
	transport := &http.Transport{
		DisableCompression: true,
	}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return transport
}

func NewMirror(cfg *conf.Mirror, logger log.Logger) (*Mirror, error) {
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 30,
	}
	if cfg.Timeout.AsDuration() > 0 {
		client.Timeout = cfg.Timeout.AsDuration()
	}
	client.Transport = makeTransport(cfg)
	if cfg.H2 {
		client.Transport = makeH2Transport(cfg)
	}

	baseURL := &url.URL{
		Scheme: "https",
		Host:   cfg.Server,
	}
	if cfg.WithoutTls {
		baseURL.Scheme = "http"
	}
	m := &Mirror{
		baseURL: *baseURL,
		client:  client,
		logger:  logger,
	}
	return m, nil
}

func (m *Mirror) urlFor(path string) string {
	builder := m.baseURL
	builder.Path = path
	return builder.String()
}

func (m *Mirror) handler(inReq *http.Request) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(inReq.Context(), inReq.Method, m.urlFor(inReq.URL.EscapedPath()), nil)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	for key, vals := range inReq.Header {
		req.Header[key] = slices.Clone(vals)
	}
	reqBody, err := ioutil.ReadAll(inReq.Body)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	out := &http.Response{}
	out.Header = http.Header{}
	for key, vals := range resp.Header {
		out.Header[key] = slices.Clone(vals)
	}
	out.StatusCode = resp.StatusCode
	out.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return out, body, nil
}

func (m *Mirror) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp, body, err := m.handler(req)
	if err != nil {
		log.NewHelper(m.logger).Errorf("failed to handling request: %q: %+v", req.URL.String(), err)
		return
	}

	switch {
	case http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices:
		log.NewHelper(m.logger).Infof("handling incoming request: %q, status-code: %d, body-size: %d", req.URL.String(), resp.StatusCode, len(body))
	default:
		log.NewHelper(m.logger).Warnf("handling incoming request with failed status code: %q, status-code: %d, body: %s", req.URL.String(), resp.StatusCode, string(body))
	}

	for key, vals := range resp.Header {
		w.Header()[key] = slices.Clone(vals)
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(body); err != nil {
		log.NewHelper(m.logger).Errorf("failed to write response body: %q: %+v", req.URL.String(), err)
	}
}
