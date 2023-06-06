package http

import (
	"bufio"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var proxyURL *url.URL
var defaultHttpTimeout = 20 * time.Second
var proxyChanged bool

var Client = &http2{}

type http2 struct {
	isAnotherRequest bool
	client           *http.Client
	mutex            sync.Mutex
	initialized      bool
}

func (ch *http2) NewHttpClient(timeout ...time.Duration) *http.Client {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if len(timeout) == 1 {
		return &http.Client{
			Transport: transCfg,   // disable tls verify
			Timeout:   timeout[0], //必须设置一个超时，不然程序会抛出非自定义错误
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}, //不跳转302
		}
	}
	return &http.Client{
		Transport: transCfg,           // disable tls verify
		Timeout:   defaultHttpTimeout, //必须设置一个超时，不然程序会抛出非自定义错误
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (ch *http2) Do(req *http.Request) (*http.Response, error) {
	if ch.client == nil {
		ch.client = ch.NewHttpClient()
	}
	ch.mutex.Lock()
	if proxyChanged {
		ch.client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
		proxyChanged = false
	}
	ch.mutex.Unlock()
	return ch.client.Do(req)
}

func GetResponseBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	reader := bufio.NewReader(body)
	var result []byte
	for {
		buffer := make([]byte, 1024)
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		result = append(result, buffer[:n]...)
	}
	return result, nil
}

func SetGlobalProxy(proxyServer string) error {
	if strings.TrimSpace(proxyServer) == "" {
		proxyURL = nil
		proxyChanged = true
		return nil
	}
	newProxyURL, err := url.Parse(proxyServer)
	if err != nil {
		return err
	}
	proxyURL = newProxyURL
	proxyChanged = true
	return nil
}
