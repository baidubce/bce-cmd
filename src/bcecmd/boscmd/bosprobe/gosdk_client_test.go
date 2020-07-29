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
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
)

var (
	bosClient *bos.Client
)

type GoSdkFake struct {
	cli bce.Client
}

func NewGoSdkFake(cli bce.Client) *GoSdkFake {
	return &GoSdkFake{cli: cli}
}

// Put an object from string
func (g *GoSdkFake) putObjectFromString(bucket, object, content string) (*serviceResp, error) {
	probeResp := &serviceResp{}
	probeResp.requestId = bucket
	probeResp.debugId = object
	probeResp.status = len(content)

	if bucket == "liupeng-bj" || bucket == "liupeng-gz" {
		return probeResp, nil
	} else if bucket == "serverErr" {
		return probeResp, &bce.BceServiceError{
			Code: boscmd.CODE_INVALID_BUCKET_NAME,
		}
	}
	return nil, &bce.BceClientError{
		Message: "ClientErr",
	}
}

// Put an object from file
func (g *GoSdkFake) putObjectFromFile(bucket, object, fileName string) (*serviceResp, error) {
	probeResp := &serviceResp{}
	probeResp.requestId = bucket
	probeResp.debugId = object

	if bucket == "liupeng-bj" || bucket == "liupeng-gz" {
		return probeResp, nil
	} else if bucket == "serverErr" {
		return probeResp, &bce.BceServiceError{
			Code: boscmd.CODE_INVALID_BUCKET_NAME,
		}
	}
	return nil, &bce.BceClientError{
		Message: "ClientErr",
	}
}

// get the name of the first object in bucket
func (g *GoSdkFake) getOneObjectFromBucket(bucket string) (*serviceResp, string, error) {
	probeResp := &serviceResp{}
	probeResp.requestId = bucket

	if bucket == "liupeng-bj" || bucket == "liupeng-gz" {
		return probeResp, "firstObject", nil
	} else if bucket == "serverErr" {
		return probeResp, "", &bce.BceServiceError{
			Code: boscmd.CODE_INVALID_BUCKET_NAME,
		}
	}
	return nil, "", &bce.BceClientError{
		Message: "ClientErr",
	}
	return nil, bucket, nil
}

// Get object, save to localpath
func (g *GoSdkFake) getObject(bucket, object, localPath string) (*serviceResp, error) {

	probeResp := &serviceResp{}
	probeResp.requestId = bucket
	probeResp.debugId = object

	if bucket == "liupeng-bj" || bucket == "liupeng-gz" {
		return probeResp, nil
	} else if bucket == "serverErr" {
		return probeResp, &bce.BceServiceError{
			Code: boscmd.CODE_INVALID_BUCKET_NAME,
		}
	}
	return nil, &bce.BceClientError{
		Message: "ClientErr",
	}
}

// Get object from given url
func (g *GoSdkFake) getObjectFromUrl(url, localPath string) (*serviceResp, error) {
	probeResp := &serviceResp{}
	probeResp.requestId = url
	probeResp.debugId = localPath
	return probeResp, nil
}
