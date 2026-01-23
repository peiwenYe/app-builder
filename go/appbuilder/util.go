// Copyright (c) 2024 Baidu, Inc. All Rights Reserved.
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

package appbuilder

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type SSEEvent struct {
	LastEventID string
	Type        string
	Data        string
}

func checkHTTPResponse(rsp *http.Response) (string, error) {
	requestID := rsp.Header.Get("X-Appbuilder-Request-Id")
	if rsp.StatusCode == http.StatusOK {
		log.Printf("Successful HTTP response. RequestID: %s", requestID)
		return requestID, nil
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.Printf("Failed to read response body. RequestID: %s, Error: %v", requestID, err)
		return requestID, err
	}
	log.Printf("HTTP response with unexpected status code. RequestID: %s, StatusCode: %d, Content: %s", requestID, rsp.StatusCode, string(data))
	return requestID, fmt.Errorf("http status code is %d, content is %s", rsp.StatusCode, string(data))
}

func NewSSEReader(bufSize int, reader *bufio.Reader) *sseReader {
	//buf := make([]byte, bufSize)
	return &sseReader{reader: reader, buf: bytes.Buffer{}}
}

type sseReader struct {
	reader *bufio.Reader
	buf    bytes.Buffer
}

func (t *sseReader) ReadMessageLine() ([]byte, error) {
	t.buf.Reset()
	for {
		line, isPrefix, err := t.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		t.buf.Grow(len(line))
		t.buf.Write(line)
		if !isPrefix {
			break
		}
	}
	// 读取空行
	line, _, err := t.reader.ReadLine()
	if err != nil || len(line) != 0 {
		t.buf.Grow(len(line))
		return nil, errors.New(t.buf.String())
	}
	return t.buf.Bytes(), nil
}
