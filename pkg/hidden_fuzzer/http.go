package hidden_fuzzer

import (
	"io"
	"net/http"
	"time"
)

// TO DO make runable with POST method from user interface
func sendHTTPRequest(client http.Client, request Request) (*Response, error) {
	start := time.Now()
	req, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	//client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//bodySmilar := fmt.Sprint(resp.Header) + string(data) # this usage removed for api fuzzing
	bodySmilar := string(data)

	response := &Response{
		URL:            request.URL,
		StatusCode:     resp.StatusCode,
		Headers:        resp.Header,
		Body:           string(data),
		ContentLength:  resp.ContentLength,
		ContentType:    resp.Header.Get("Content-Type"),
		Time:           time.Since(start),
		DataForSimilar: bodySmilar,
	}

	return response, nil
}
