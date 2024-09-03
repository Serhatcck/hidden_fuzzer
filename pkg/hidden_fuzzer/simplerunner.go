package hidden_fuzzer

import (
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Download results < 5MB
const MAX_DOWNLOAD_SIZE = 5242880

type SimpleRunner struct {
	config *Config
	client *http.Client
}

func NewSimpleRunner(conf *Config, replay bool) *SimpleRunner {
	var simplerunner SimpleRunner
	proxyURL := http.ProxyFromEnvironment
	customProxy := ""

	/*if replay {
		customProxy = conf.ReplayProxyURL
	} else {
		customProxy = conf.ProxyURL
	}*/
	if len(customProxy) > 0 {
		pu, err := url.Parse(customProxy)
		if err == nil {
			proxyURL = http.ProxyURL(pu)
		}
	}
	cert := []tls.Certificate{}

	/*if conf.ClientCert != "" && conf.ClientKey != "" {
		tmp, _ := tls.LoadX509KeyPair(conf.ClientCert, conf.ClientKey)
		cert = []tls.Certificate{tmp}
	}*/

	simplerunner.config = conf
	simplerunner.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:       time.Duration(time.Duration(10) * time.Second),
		Transport: &http.Transport{
			//ForceAttemptHTTP2:   conf.Http2,
			Proxy:               proxyURL,
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			MaxConnsPerHost:     200,
			DialContext: (&net.Dialer{
				Timeout: time.Duration(time.Duration(10) * time.Second),
			}).DialContext,
			TLSHandshakeTimeout: time.Duration(time.Duration(10) * time.Second),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
				Renegotiation:      tls.RenegotiateOnceAsClient,
				Certificates:       cert,
			},
		}}

	return &simplerunner
}

func (r *SimpleRunner) Execute(req *Request) (Response, error) {
	var httpreq *http.Request
	var err error
	//var rawreq []byte
	//data := bytes.NewReader(req.Data)

	var start time.Time
	var firstByteTime time.Duration

	/*trace := &httptrace.ClientTrace{
		WroteRequest: func(wri httptrace.WroteRequestInfo) {
			start = time.Now() // begin the timer after the request is fully written
		},
		GotFirstResponseByte: func() {
			firstByteTime = time.Since(start) // record when the first byte of the response was received
		},
	}*/

	//httpreq, err = http.NewRequestWithContext(r.config.Context, req.Method, req.URL, data)
	httpreq, err = http.NewRequest(req.Method, req.URL, nil)

	if err != nil {
		return Response{}, err
	}

	// set default User-Agent header if not present
	if _, ok := req.Headers["User-Agent"]; !ok {
		req.Headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	}

	// Handle Go http.Request special cases
	if _, ok := req.Headers["Host"]; !ok {
		req.Headers["Host"] = req.Host
	}

	req.Host = httpreq.URL.Hostname()
	req.Schema = httpreq.URL.Scheme
	//httpreq = httpreq.WithContext(httptrace.WithClientTrace(r.config.Context, trace))

	for k, v := range req.Headers {
		httpreq.Header.Set(k, v)
	}

	httpresp, err := r.client.Do(httpreq)
	if err != nil {
		return Response{}, err
	}
	//resp := ffuf.NewResponse(httpresp, req)
	resp := &Response{
		URL:        req.URL,
		StatusCode: httpresp.StatusCode,
		Headers:    httpresp.Header,
		//Body:           string(data),
		ContentLength: httpresp.ContentLength,
		ContentType:   httpresp.Header.Get("Content-Type"),
		Time:          time.Since(start),
		//DataForSimilar: bodySmilar,
		Request: *req,
	}
	defer httpresp.Body.Close()

	// Check if we should download the resource or not
	size, err := strconv.Atoi(httpresp.Header.Get("Content-Length"))
	if err == nil {
		resp.ContentLength = int64(size)
		if size > MAX_DOWNLOAD_SIZE {
			return *resp, nil
		}
	}

	var bodyReader io.ReadCloser
	if httpresp.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(httpresp.Body)
		if err != nil {
			// fallback to raw data
			bodyReader = httpresp.Body
		}
	} else if httpresp.Header.Get("Content-Encoding") == "deflate" {
		bodyReader = flate.NewReader(httpresp.Body)
		if err != nil {
			// fallback to raw data
			bodyReader = httpresp.Body
		}
	} else {
		bodyReader = httpresp.Body
	}

	if respbody, err := io.ReadAll(bodyReader); err == nil {
		resp.ContentLength = int64(len(string(respbody)))
		resp.Body = string(respbody)
	}

	resp.Time = firstByteTime
	return *resp, nil
}
