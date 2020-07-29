// Copyright 2017 Baidu, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package bosprobe

import (
	"encoding/json"
	"fmt"
	"github.com/baidubce/bce-sdk-go/util/log"
	"io"
	nethttp "net/http"
	"os"
	"strconv"
)

import (
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/http"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

type serviceResp struct {
	status    int
	requestId string
	debugId   string
}

// define gosdk interface for easier to test
type goSdk interface {
	putObjectFromString(bucket, object, content string) (*serviceResp, error)
	putObjectFromFile(bucket, object, fileName string) (*serviceResp, error)
	getOneObjectFromBucket(bucket string) (*serviceResp, string, error)
	getObject(bucket, object, localPath string) (*serviceResp, error)
	getObjectFromUrl(url, localPath string) (*serviceResp, error)
}

// Get a new bos gosdk client
func NewClient(ak, sk, endpoint string) (*bos.Client, error) {
	bosClient, err := bos.NewClient(ak, sk, endpoint)
	if err != nil {
		return nil, err
	}
	bosClient.Config.Retry = bce.NewBackOffRetryPolicy(0, 0, 0)
	bosClient.Config.UserAgent = PROBE_AGENT
	return bosClient, nil
}

func NewGoSdk(cli bce.Client) goSdk {
	return NewGoSdkImp(cli)
}

type GoSdkImp struct {
	cli bce.Client
}

func NewGoSdkImp(cli bce.Client) *GoSdkImp {
	return &GoSdkImp{cli: cli}
}

// Put an object from string
func (g *GoSdkImp) putObjectFromString(bucket, object, content string) (*serviceResp, error) {
	log.Debugf("read from file")
	body, err := bce.NewBodyFromString(content)
	log.Debugf("read file finish")
	if err != nil {
		return nil, err
	}
	return g.putObject(bucket, object, body)
}

// Put an object from file
func (g *GoSdkImp) putObjectFromFile(bucket, object, fileName string) (*serviceResp, error) {
	body, err := bce.NewBodyFromFile(fileName)
	if err != nil {
		return nil, err
	}
	return g.putObject(bucket, object, body)
}

// Put object imp
func (g *GoSdkImp) putObject(bucket, object string, body *bce.Body) (*serviceResp, error) {
	req := &bce.BceRequest{}
	req.SetUri(bce.URI_PREFIX + bucket + "/" + object)
	req.SetMethod(http.PUT)
	req.SetBody(body)
	resp := &bce.BceResponse{}

	err := g.cli.SendRequest(req, resp)
	if err != nil {
		if _, ok := err.(*bce.BceServiceError); !ok {
			return nil, err
		}
	}

	probeResp := &serviceResp{}
	probeResp.status = resp.StatusCode()
	probeResp.requestId = resp.RequestId()
	probeResp.debugId = resp.DebugId()

	if resp.IsFail() {
		err = resp.ServiceError()
	}
	return probeResp, err
}

// get the name of the first object in bucket
func (g *GoSdkImp) getOneObjectFromBucket(bucket string) (*serviceResp, string, error) {
	var (
		objectName string
		err        error
	)
	req := &bce.BceRequest{}
	req.SetUri(bce.URI_PREFIX + bucket)
	req.SetMethod(http.GET)
	req.SetParam("maxKeys", "1")
	resp := &bce.BceResponse{}

	err = g.cli.SendRequest(req, resp)
	if err != nil {
		if _, ok := err.(*bce.BceServiceError); !ok {
			return nil, "", err
		}
	}

	probeResp := &serviceResp{}
	probeResp.status = resp.StatusCode()
	probeResp.requestId = resp.RequestId()
	probeResp.debugId = resp.DebugId()

	if !resp.IsFail() {
		result := &api.ListObjectsResult{}
		if err := resp.ParseJsonBody(result); err != nil {
			return probeResp, "", err
		}
		if len(result.Contents) > 0 {
			objectName = result.Contents[0].Key
		} else {
			return probeResp, "", &bce.BceServiceError{
				Code:    SERVER_RETURN_EMPTY_OBJECT_LIST,
				Message: "Bos return a empty object list",
			}
		}
	} else {
		err = resp.ServiceError()
		serverErr, ok := err.(*bce.BceServiceError)
		if !ok {
			return probeResp, "", err
		}
		// user maybe not have right to list bucket, but he have right to read
		if serverErr.Code == bce.EACCESS_DENIED {
			return probeResp, "", &bce.BceServiceError{
				Code:    SERVER_GET_OBJECT_LIST_DENIED,
				Message: serverErr.Message,
			}
		}
	}
	return probeResp, objectName, err
}

// Get object, save to localpath
func (g *GoSdkImp) getObject(bucket, object, localPath string) (*serviceResp, error) {
	var (
		contentLength string
	)

	req := &bce.BceRequest{}
	req.SetUri(bce.URI_PREFIX + bucket + "/" + object)
	req.SetMethod(http.GET)

	// Send request and get the result
	resp := &bce.BceResponse{}
	err := g.cli.SendRequest(req, resp)
	if err != nil {
		if _, ok := err.(*bce.BceServiceError); !ok {
			return nil, err
		}
	}

	probeResp := &serviceResp{}
	probeResp.status = resp.StatusCode()
	probeResp.requestId = resp.RequestId()
	probeResp.debugId = resp.DebugId()

	if resp.IsFail() {
		return probeResp, resp.ServiceError()
	}

	body := resp.Body()
	defer body.Close()
	if val, ok := resp.Headers()[http.CONTENT_LENGTH]; ok {
		contentLength = val
	}
	err = g.writeBodyToFile(localPath, contentLength, body)
	if err != nil {
		return probeResp, err
	}
	return probeResp, nil
}

// Get object from given url
func (g *GoSdkImp) getObjectFromUrl(url, localPath string) (*serviceResp, error) {
	var (
		contentLength string
	)
	httpResp, err := nethttp.Get(url)

	if err != nil {
		return nil, &bce.BceClientError{"execute http request failed!"}
	}

	probeResp := &serviceResp{}
	probeResp.status = httpResp.StatusCode
	probeResp.requestId = httpResp.Header.Get(http.BCE_REQUEST_ID)
	probeResp.debugId = httpResp.Header.Get(http.BCE_DEBUG_ID)

	if httpResp.StatusCode >= 400 {
		serverErr := &bce.BceServiceError{}
		defer httpResp.Body.Close()
		jsonDecoder := json.NewDecoder(httpResp.Body)
		decodeOk := jsonDecoder.Decode(serverErr)
		if decodeOk != nil {
			serverErr.Code = bce.EMALFORMED_JSON
			serverErr.Message = "Service json error message decode failed"
			serverErr.RequestId = probeResp.requestId
			serverErr.StatusCode = probeResp.status
			return probeResp, serverErr
		}
		return probeResp, serverErr
	}

	if val := httpResp.Header.Get(http.CONTENT_LENGTH); val != "" {
		contentLength = val
	}

	body := httpResp.Body
	defer body.Close()

	err = g.writeBodyToFile(localPath, contentLength, body)
	if err != nil {
		return probeResp, err
	}
	return probeResp, nil
}

// Write stream to file
func (g *GoSdkImp) writeBodyToFile(localPath string, contentLength string,
	body io.ReadCloser) error {

	file, fileErr := os.OpenFile(localPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	size, lenErr := strconv.ParseInt(contentLength, 10, 64)
	if lenErr != nil {
		return lenErr
	}
	written, writeErr := io.CopyN(file, body, size)
	if writeErr != nil {
		return writeErr
	}
	if written != size {
		return fmt.Errorf("written content size does not match the response content")
	}
	return nil
}
