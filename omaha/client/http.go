// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/coreos/go-omaha/omaha"
)

const (
	defaultTimeout = 90 * time.Second
	defaultTries   = 7
)

// httpClient extends the standard http.Client to support xml encoding
// and decoding as well as automatic retries on transient failures.
type httpClient struct {
	http.Client
}

func newHTTPClient() *httpClient {
	return &httpClient{http.Client{
		Timeout: defaultTimeout,
	}}
}

// doPost sends a single HTTP POST, returning a parsed omaha response.
func (hc *httpClient) doPost(url string, reqBody []byte) (*omaha.Response, error) {
	resp, err := hc.Post(url, "text/xml; charset=utf-8", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	omahaResp, err := omaha.ParseResponse(contentType, resp.Body)

	// Prefer reporting HTTP errors over XML parsing errors.
	if resp.StatusCode != http.StatusOK {
		err = &httpError{resp}
	}

	return omahaResp, err
}

// Omaha encodes and sends an omaha request, retrying on any transient errors.
func (hc *httpClient) Omaha(url string, req *omaha.Request) (resp *omaha.Response, err error) {
	buf := bytes.NewBufferString(xml.Header)
	enc := xml.NewEncoder(buf)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("omaha: failed to encode request: %v", err)
	}

	for i := 0; i < defaultTries; i++ {
		resp, err = hc.doPost(url, buf.Bytes())
		if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
			// TODO(marineam): add exponential backoff
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("omaha: request failed: %v", err)
	}

	return resp, nil
}
