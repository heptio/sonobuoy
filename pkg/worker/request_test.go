/*
Copyright 2018 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package worker

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

func TestErrorRequestRetry(t *testing.T) {

	tests := []struct {
		name string
		f    func() (io.Reader, string, error)
	}{
		{
			name: "error request retry",
			f: func() (io.Reader, string, error) {
				return nil, "", errors.New("didn't succeed")
			},
		},
		{
			name: "success request retry",
			f: func() (io.Reader, string, error) {
				return bytes.NewBuffer([]byte("success!")), "success!", nil
			},
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			testServer := &testServer{
				responseCodes: []int{500, 200},
			}

			server := httptest.NewTLSServer(testServer)
			defer server.Close()

			err := DoRequest(server.URL, server.Client(), test.f)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if testServer.responseCount != 2 {
				t.Errorf("expected 2 requests, got %d", testServer.responseCount)
			}
		})
	}
}

type testServer struct {
	sync.Mutex
	responseCodes []int
	responseCount int
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.Lock()
	defer t.Unlock()

	responseCode := 500

	if len(t.responseCodes) > 0 {
		responseCode, t.responseCodes = t.responseCodes[0], t.responseCodes[1:]
	}

	w.WriteHeader(responseCode)
	w.Write([]byte("ok!"))

	t.responseCount++
}
