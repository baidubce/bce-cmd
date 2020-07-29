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

package boscli

import (
	"fmt"
	nethttp "net/http"
	"os"
	// 	"runtime"
	"strings"
	"testing"
	// 	"path/filepath"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"utils/util"
)

var (
	testBosClient *bosClientWrapper
	orgBosClient  *bos.Client
	bucket_bj     = "cli-test"
	bucket_bj1    = "cli-test-bj"
	bucket_gz     = "cli-test-gz"
	bucket_su     = "cli-test-su"
	cli_test      = "cli-test"
	bucket_error  = "cli-test_xxx"
	bucketDomains = map[string]string{
		bucket_bj: "bj.bcebos.com",
		bucket_gz: "gz.bcebos.com",
		bucket_su: "su.bcebos.com",
		cli_test:  "bj.bcebos.com",
	}
	lifecycleTemp = `{
		"rule": [
		{
			 "status": "enabled",
	 		 "action": {
	 				"name": "Transition",
	 				"storageClass": "STANDARD_IA"
	 			},
	 			"resource": [
	 					"%s/*"
	 			],
	 			"id": "sample-rule-transition-1",
	 			"condition": {
	 				"time": {
	 					"dateGreaterThan": "$(lastModified)+P180D"
					}
				},
				"test":"yes"
		  },
		  {
			 "status": "enabled",
	 		 "action": {
	 				"name": "Transition",
	 				"storageClass": "COLD"
	 			},
	 			"resource": [
	 					"%s/*"
	 			],
	 			"id": "sample-rule-transition-2",
	 			"condition": {
	 				"time": {
	 					"%s": "$(lastModified)+P300D"
					}
				}
		  }
		]
		}`
)

func init() {
	var (
		err error
	)

	if err := initConfig(); err != nil {
		os.Exit(1)
	}

	orgBosClient, err = buildBosClient("", "", "", bceconf.CredentialProvider,
		bceconf.ServerConfigProvider)
	if err != nil {
		fmt.Printf("init testBosClient failed")
		os.Exit(1)
	}

	testBosClient = &bosClientWrapper{bosClient: orgBosClient}
}

type retryHandlerType struct {
	bucket    string
	endpoint  string
	cacheHave bool
	bosHave   bool
	needRetry int
	isSuc     bool
	err       error
}

type retryHandlerTestReq struct {
	bucket      string
	retryRecord int
	needRetry   int
	err         error
}

func (r *retryHandlerTestReq) getBucketName() string {
	return r.bucket
}

type retryHandlerTestResp struct {
	endpoint string
}

func TestRetryHandler(t *testing.T) {
	testCases := []retryHandlerType{
		retryHandlerType{
			bucket:    bucket_bj,
			endpoint:  "bj.bcebos.com",
			cacheHave: true,
			needRetry: 0,
			isSuc:     true,
		},
		retryHandlerType{
			bucket:    bucket_bj,
			endpoint:  "bj.bcebos.com",
			cacheHave: true,
			needRetry: 1,
			err:       &bce.BceServiceError{Code: boscmd.CODE_NO_SUCH_BUCKET},
			isSuc:     true,
		},
		retryHandlerType{
			bucket:    bucket_bj,
			endpoint:  "bj.bcebos.com",
			cacheHave: true,
			needRetry: 1,
			err:       &bce.BceServiceError{Code: boscmd.CODE_INVALID_ARGUMENT},
			isSuc:     false,
		},

		retryHandlerType{
			bucket:    bucket_error,
			endpoint:  "bj.bcebos.com",
			cacheHave: true,
			needRetry: 3,
			err:       &bce.BceServiceError{Code: boscmd.CODE_NO_SUCH_BUCKET},
			isSuc:     false,
		},

		retryHandlerType{
			bucket:    bucket_bj,
			endpoint:  "bj.bcebos.com",
			needRetry: 1,
			err:       &bce.BceServiceError{Code: boscmd.CODE_NO_SUCH_BUCKET},
			isSuc:     false,
		},
		retryHandlerType{
			bucket:    bucket_error,
			needRetry: 2,
			err:       &bce.BceServiceError{Code: boscmd.CODE_NO_SUCH_BUCKET},
			isSuc:     false,
		},
	}

	testFunc := func(bosClient *bos.Client, req boscliReq, rsp interface{}) error {

		tReq, ok := req.(*retryHandlerTestReq)
		if !ok {
			return fmt.Errorf("Error TestRetryHandler request type!")
		}

		if tReq.retryRecord < tReq.needRetry {
			tReq.retryRecord += 1
			return tReq.err
		}
		tRsp, ok := rsp.(*retryHandlerTestResp)
		if !ok {
			return fmt.Errorf("Error TestRetryHandler response type!")
		}
		tRsp.endpoint = bosClient.Config.Endpoint
		return nil
	}

	bosClient, err := buildBosClient("", "", "", bceconf.CredentialProvider,
		bceconf.ServerConfigProvider)
	if err != nil {
		t.Errorf("init bosClient failed")
		return
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint, 1080)
			if !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}
		req := &retryHandlerTestReq{
			bucket:    tCase.bucket,
			needRetry: tCase.needRetry,
			err:       tCase.err,
		}
		rsp := &retryHandlerTestResp{}
		err := retryHandler(bosClient, testFunc, req, rsp)

		util.ExpectEqual("wraper.go TestRetryHandler I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("wraper.go TestRetryHandler II", i+1, t.Errorf, true,
				strings.HasSuffix(rsp.endpoint, tCase.endpoint))
			if err != nil {
				t.Logf("error: %s", err)
			}
		}
	}
}

type headBucketType struct {
	bucket string
	isSuc  bool
}

func TestHeadBucket(t *testing.T) {
	testCases := []headBucketType{
		headBucketType{
			bucket: bucket_bj,
			isSuc:  true,
		},
		headBucketType{
			bucket: bucket_error,
			isSuc:  false,
		},

		headBucketType{
			bucket: "",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		err := testBosClient.HeadBucket(tCase.bucket)
		util.ExpectEqual("wraper.go HeadBucket I", i+1, t.Errorf, tCase.isSuc, err == nil)
	}
}

func TestListBuckets(t *testing.T) {
	_, err1 := testBosClient.ListBuckets()
	_, err2 := testBosClient.bosClient.ListBuckets()
	util.ExpectEqual("wraper.go ListBuckets I", 1, t.Errorf, err1, err2)
}

type getBucketLocationType struct {
	bucket string
	local  string
	isSuc  bool
}

func TestGetBucketLocation(t *testing.T) {
	testCases := []getBucketLocationType{
		getBucketLocationType{
			bucket: bucket_bj,
			local:  "bj",
			isSuc:  true,
		},
		getBucketLocationType{
			bucket: bucket_su,
			local:  "su",
			isSuc:  true,
		},
		getBucketLocationType{
			bucket: bucket_error,
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		ret, err := testBosClient.GetBucketLocation(tCase.bucket)
		if tCase.isSuc {
			util.ExpectEqual("wraper.go GetBucketLocation I", i+1, t.Errorf, tCase.local, ret)
		} else {
			util.ExpectEqual("wraper.go GetBucketLocation II", i+1, t.Errorf, false, err == nil)
		}
	}
}

type listObjectsType struct {
	bucket    string
	objectKey string
	endpoint  string
	cacheHave bool
	Delimiter string
	Marker    string
	MaxKeys   int
	Prefix    string
	isSuc     bool
}

func TestListObjects(t *testing.T) {
	testCases := []listObjectsType{
		listObjectsType{
			bucket:    bucket_bj,
			endpoint:  "bj.bcebos.com",
			cacheHave: true,
			Delimiter: "/",
			Marker:    "",
			MaxKeys:   10,
			Prefix:    "",
			isSuc:     true,
		},
		listObjectsType{
			bucket:    bucket_bj,
			endpoint:  "gz.bcebos.com",
			cacheHave: true,
			Delimiter: "/",
			Marker:    "",
			MaxKeys:   10,
			Prefix:    "",
			isSuc:     true,
		},
		listObjectsType{
			bucket:    bucket_error,
			Delimiter: "/",
			Marker:    "",
			MaxKeys:   10,
			Prefix:    "",
			isSuc:     false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		args := &api.ListObjectsArgs{
			Delimiter: tCase.Delimiter,
			Marker:    tCase.Marker,
			MaxKeys:   tCase.MaxKeys,
			Prefix:    tCase.objectKey,
		}
		ret1, err1 := testBosClient.ListObjects(tCase.bucket, args)
		ret2, err2 := orgBosClient.ListObjects(tCase.bucket, args)
		util.ExpectEqual("wraper.go ListObjects I", i+1, t.Errorf, err1 == nil, err2 == nil)
		if err1 == nil && err2 == nil {
			util.ExpectEqual("wraper.go ListObjects II", i+1, t.Errorf, ret1, ret2)
		}
	}
}

type putAndDeleteBucket struct {
	bucket   string
	endpoint string
	isSuc    bool
}

func TestPutAndDeleteBucket(t *testing.T) {
	testCases := []putAndDeleteBucket{
		putAndDeleteBucket{
			bucket:   "cli-test-xxx",
			endpoint: "bj.bcebos.com",
			isSuc:    true,
		},
		putAndDeleteBucket{
			bucket:   bucket_bj,
			endpoint: "bj.bcebos.com",
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		var (
			delErr1 error
			delErr2 error
		)
		_, putErr1 := testBosClient.PutBucket(tCase.bucket)
		if putErr1 == nil {
			delErr1 = testBosClient.DeleteBucket(tCase.bucket)
		} else {
			t.Logf("put bucket error1: %s", putErr1)
		}
		_, putErr2 := orgBosClient.PutBucket(tCase.bucket)
		if putErr2 == nil {
			delErr2 = orgBosClient.DeleteBucket(tCase.bucket)
		} else {
			t.Logf("put bucket error2: %s", putErr2)
		}
		util.ExpectEqual("wraper.go putBucket I", i+1, t.Errorf, putErr1 == nil, putErr2 == nil)
		util.ExpectEqual("wraper.go deleteBucket I", i+1, t.Errorf, delErr1 == nil, delErr2 == nil)
	}
}

// Get object from given url
func canGetObjectFromUrl(url string) bool {
	httpResp, err := nethttp.Get(url)

	if err != nil {
		return false
	}

	if httpResp.StatusCode >= 400 {
		return false
	} else {
		return true
	}
}

type genUrlType struct {
	bucket string
	object string
	expire int
	isSuc  bool
	wait   int
}

func TestGenUrl(t *testing.T) {
	testCases := []genUrlType{
		genUrlType{
			bucket: "cli-test-xxx",
			object: "123",
			isSuc:  false,
		},
		genUrlType{
			bucket: cli_test,
			object: "bce",
			isSuc:  true,
		},
	}
	for i, tCase := range testCases {
		ret := testBosClient.BasicGeneratePresignedUrl(tCase.bucket, tCase.object, 100)
		util.ExpectEqual("wraper.go putBucket I", i+1, t.Errorf, tCase.isSuc,
			canGetObjectFromUrl(ret))
	}
}

type delObjectsType struct {
	bucket    string
	keyList   []string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestDelObjects(t *testing.T) {
	testCases := []delObjectsType{
		delObjectsType{
			bucket:    cli_test,
			keyList:   []string{"te1234", "tel456"},
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     false,
		},
		delObjectsType{
			bucket:  cli_test,
			keyList: []string{"te1234", "tel456"},
			isSuc:   false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		ret1, err1 := testBosClient.DeleteMultipleObjectsFromKeyList(tCase.bucket, tCase.keyList)
		ret2, err2 := orgBosClient.DeleteMultipleObjectsFromKeyList(tCase.bucket, tCase.keyList)
		util.ExpectEqual("wraper.go DeleteMultipleObjectsFromKeyList I", i+1, t.Errorf, err1 == nil,
			err2 == nil)
		if err1 == nil && err2 == nil {
			util.ExpectEqual("wraper.go DeleteMultipleObjectsFromKeyList II", i+1, t.Errorf,
				len(ret1.Errors), len(ret2.Errors))
		}
	}
}

type delObjectType struct {
	bucket    string
	object    string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestDelObject(t *testing.T) {
	testCases := []delObjectType{
		delObjectType{
			bucket:    cli_test,
			object:    "te1234",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     false,
		},
		delObjectType{
			bucket: cli_test,
			object: "te1234",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		err1 := testBosClient.DeleteObject(tCase.bucket, tCase.object)
		err2 := orgBosClient.DeleteObject(tCase.bucket, tCase.object)
		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go DeleteObject I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go DeleteObject II", i+1, t.Errorf, serverErr1.Code,
					serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go DeleteObject III", i+1, t.Errorf, err1, err2)
		}
	}
}

type getObjectMetaType struct {
	bucket    string
	object    string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestGetObjectMeta(t *testing.T) {
	testCases := []getObjectMetaType{
		getObjectMetaType{
			bucket:    cli_test,
			object:    "Readme.txt",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		getObjectMetaType{
			bucket: cli_test,
			object: "Readme.txt",
			isSuc:  true,
		},
		getObjectMetaType{
			bucket: cli_test,
			object: "te1234",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		ret1, err1 := testBosClient.GetObjectMeta(tCase.bucket, tCase.object)
		ret2, err2 := orgBosClient.GetObjectMeta(tCase.bucket, tCase.object)

		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go getObjectMeta I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go getObjectMeta II", i+1, t.Errorf, serverErr1.Code,
					serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go getObjectMeta III", i+1, t.Errorf, err1, err2)
			if err1 == nil {
				util.ExpectEqual("wraper.go getObjectMeta IV", i+1, t.Errorf, ret1.ETag, ret2.ETag)
			}
		}
	}
}

type copyObjectType struct {
	bucket       string
	object       string
	srcBucket    string
	srcObject    string
	storageClass string
	cacheHave    bool
	endpoint     string
	isSuc        bool
}

func TestCopyObject(t *testing.T) {
	testCases := []copyObjectType{
		copyObjectType{
			srcBucket: cli_test,
			srcObject: "bce",
			bucket:    bucket_bj1,
			object:    "bce",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		copyObjectType{
			srcBucket: cli_test,
			srcObject: "bce",
			bucket:    bucket_bj1,
			object:    "bce",
			isSuc:     true,
		},
		copyObjectType{
			srcBucket: cli_test,
			srcObject: "te1234",
			bucket:    bucket_bj1,
			object:    "te1234",
			isSuc:     false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		ret1, err1 := testBosClient.CopyObject(tCase.bucket, tCase.object, tCase.srcBucket,
			tCase.srcObject, nil)
		ret2, err2 := orgBosClient.CopyObject(tCase.bucket, tCase.object, tCase.srcBucket,
			tCase.srcObject, nil)

		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go CopyObject I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go CopyObject II", i+1, t.Errorf, serverErr1.Code,
					serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go CopyObject III", i+1, t.Errorf, err1, err2)
			if err1 == nil && err2 == nil {
				util.ExpectEqual("wraper.go CopyObject IV", i+1, t.Errorf, ret1.ETag, ret2.ETag)
			}
		}
	}
}

type basicGetObjectToFileType struct {
	bucket    string
	object    string
	local     string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestBasicGetObjectToFile(t *testing.T) {
	testCases := []basicGetObjectToFileType{
		basicGetObjectToFileType{
			bucket:    bucket_bj,
			object:    "bce",
			local:     "download_test",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		basicGetObjectToFileType{
			bucket: bucket_bj,
			object: "bce",
			local:  "download_test",
			isSuc:  true,
		},
		basicGetObjectToFileType{
			bucket: bucket_bj,
			object: "te1234",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		var (
			size1 int64 = -1
			size2 int64 = -2
		)
		err1 := testBosClient.BasicGetObjectToFile(tCase.bucket, tCase.object, tCase.local)
		if err1 == nil {
			size1, _ = util.GetSizeOfFile(tCase.local)
		}
		err2 := orgBosClient.BasicGetObjectToFile(tCase.bucket, tCase.object, tCase.local)
		if err2 == nil {
			size2, _ = util.GetSizeOfFile(tCase.local)
		}

		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go basicGetObjectToFile I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go basicGetObjectToFile II", i+1, t.Errorf,
					serverErr1.Code, serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go basicGetObjectToFile III", i+1, t.Errorf, err1, err2)
			if err1 == nil {
				util.ExpectEqual("wraper.go basicGetObjectToFile IV", i+1, t.Errorf, size1, size2)
			}
		}
	}
}

type putObjectFromFileType struct {
	bucket    string
	object    string
	local     string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestPutObjectFromFile(t *testing.T) {
	testCases := []putObjectFromFileType{
		putObjectFromFileType{
			bucket:    bucket_bj,
			object:    "bce",
			local:     "download_test",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		putObjectFromFileType{
			bucket: bucket_bj,
			object: "bce",
			local:  "download_test",
			isSuc:  true,
		},
		putObjectFromFileType{
			bucket: bucket_bj,
			object: "te1234",
			local:  "download_test_xxx",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		var (
			size1 int64  = -1
			size2 int64  = -2
			etag1 string = "1"
			etag2 string = "-1"
		)
		_, err1 := testBosClient.PutObjectFromFile(tCase.bucket, tCase.object, tCase.local, nil)
		if err1 == nil {
			ret, err := orgBosClient.GetObjectMeta(tCase.bucket, tCase.object)
			if err == nil {
				size1 = ret.ContentLength
				etag1 = ret.ETag
			}
		}
		_, err2 := orgBosClient.PutObjectFromFile(tCase.bucket, tCase.object, tCase.local, nil)
		if err2 == nil {
			ret, err := orgBosClient.GetObjectMeta(tCase.bucket, tCase.object)
			if err == nil {
				size2 = ret.ContentLength
				etag2 = ret.ETag
			}
		}

		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go putObjectFromFile I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go putObjectFromFile II", i+1, t.Errorf,
					serverErr1.Code, serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go putObjectFromFile III", i+1, t.Errorf, err1, err2)
			if err1 == nil {
				util.ExpectEqual("wraper.go putObjectFromFile IV", i+1, t.Errorf, size1, size2)
				util.ExpectEqual("wraper.go putObjectFromFile V", i+1, t.Errorf, etag1, etag2)
				util.ExpectEqual("wraper.go putObjectFromFile VI", i+1, t.Errorf, false,
					etag1 == "")
			}
		}
	}
}

type uploadSuperFileType struct {
	bucket    string
	object    string
	local     string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestUploadSuperFile(t *testing.T) {
	// 	if err := util.CreateFileWithSize("download_test_big", 100 << 20); err != nil {
	// 		t.Errorf("create big file download_test_big")
	// 		return
	// 	}
	testCases := []uploadSuperFileType{
		uploadSuperFileType{
			bucket:    bucket_bj,
			object:    "bce",
			local:     "download_test",
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		uploadSuperFileType{
			bucket: bucket_bj,
			object: "bce",
			local:  "download_test",
			isSuc:  true,
		},
		uploadSuperFileType{
			bucket: bucket_bj,
			object: "te1234",
			local:  "download_test_xxx",
			isSuc:  false,
		},
	}
	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		var (
			size1 int64  = -1
			size2 int64  = -2
			etag1 string = "1"
			etag2 string = "-1"
		)
		err1 := testBosClient.UploadSuperFile(tCase.bucket, tCase.object, tCase.local, "")
		if err1 == nil {
			ret, err := orgBosClient.GetObjectMeta(tCase.bucket, tCase.object)
			if err == nil {
				size1 = ret.ContentLength
				etag1 = ret.ETag
			}
		}
		err2 := orgBosClient.UploadSuperFile(tCase.bucket, tCase.object, tCase.local, "")
		if err2 == nil {
			ret, err := orgBosClient.GetObjectMeta(tCase.bucket, tCase.object)
			if err == nil {
				size2 = ret.ContentLength
				etag2 = ret.ETag
			}
		}

		if serverErr1, ok := err1.(*bce.BceServiceError); ok {
			serverErr2, ok := err2.(*bce.BceServiceError)
			util.ExpectEqual("wraper.go uploadSuperFile I", i+1, t.Errorf, true, ok)
			if ok {
				util.ExpectEqual("wraper.go uploadSuperFile II", i+1, t.Errorf,
					serverErr1.Code, serverErr2.Code)
			}
		} else {
			util.ExpectEqual("wraper.go uploadSuperFile III", i+1, t.Errorf, err1, err2)
			if err1 == nil {
				t.Logf("size %d %d etag %s %s", size1, size2, etag1, etag2)
				util.ExpectEqual("wraper.go uploadSuperFile IV", i+1, t.Errorf, size1, size2)
				util.ExpectEqual("wraper.go uploadSuperFile IV", i+1, t.Errorf, false, etag1 == "")
				util.ExpectEqual("wraper.go uploadSuperFile V", i+1, t.Errorf, true,
					etag1 != etag2)
			}
		}
	}
}

type putCannedAclType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	canned    string
	isSuc     bool
}

func TestPutBucketAclFromCanned(t *testing.T) {
	var (
		err error
	)

	testCases := []putCannedAclType{
		// 1
		putCannedAclType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			canned:    "xxx",
			isSuc:     false,
		},
		// 2
		putCannedAclType{
			bucket: bucket_bj,
			canned: "xxx",
			isSuc:  false,
		},
		// 3
		putCannedAclType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			canned:    "public-read",
			isSuc:     true,
		},

		// 4
		putCannedAclType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			canned:    "public-read-write",
			isSuc:     true,
		},
		// 5
		putCannedAclType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			canned:    "private",
			isSuc:     true,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		var (
			aclWrapper *api.GetBucketAclResult
			aclOrg     *api.GetBucketAclResult
		)

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		err1 := testBosClient.PutBucketAclFromCanned(tCase.bucket, tCase.canned)
		if err1 == nil {
			aclWrapper, err = orgBosClient.GetBucketAcl(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketAclFromCanned I, ID %d : Get ACL failed", i+1)
				continue
			}
		}
		err2 := orgBosClient.PutBucketAclFromCanned(tCase.bucket, tCase.canned)
		if err2 == nil {
			aclOrg, err = orgBosClient.GetBucketAcl(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketAclFromCanned II, ID %d : Get ACL failed: %s", i+1, err)
				continue
			}
		}

		util.ExpectEqual("wraper.go PutBucketAclFromCanned III", i+1, t.Errorf, tCase.isSuc,
			err1 == nil)

		if !tCase.isSuc {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go PutBucketAclFromCanned IV", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go PutBucketAclFromCanned V", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go PutBucketAclFromCanned VI", i+1, t.Errorf, err1, err2)
			}
		} else {
			if aclWrapper == nil {
				t.Errorf("wraper.go PutBucketAclFromCanned: %d acl is nil", i+1)
			}

			util.ExpectEqual("wraper.go PutBucketAclFromCanned VI", i+1, t.Errorf, true,
				aclWrapper != nil)
			util.ExpectEqual("wraper.go PutBucketAclFromCanned VII", i+1, t.Errorf, true,
				aclOrg != nil)

			util.ExpectEqual("wraper.go PutBucketAclFromCanned VIII", i+1, t.Errorf, aclWrapper,
				aclOrg)
		}
	}
}

type putAclFromStringType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	acl       string
	isSuc     bool
}

func TestPutBucketAclFromString(t *testing.T) {
	var (
		err error
	)

	testCases := []putAclFromStringType{
		// 1
		putAclFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			acl:       "xxx",
			isSuc:     false,
		},
		// 2
		putAclFromStringType{
			bucket: bucket_bj,
			acl:    "xxx",
			isSuc:  false,
		},
		// 3
		putAclFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			acl: "{\"accessControlList\": [{\"grantee\": [{\"id\": \"*\"}],\"permission\":" +
				"[\"READ\"]}]}",
			isSuc: true,
		},

		// 4
		putAclFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			acl: "{\"accessControlList\": [{\"grantee\": [{\"id\": \"*\"}],\"permission\": " +
				"[\"READ\", \"WRITE\"]}]}",
			isSuc: true,
		},
		// 5
		putAclFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			acl: "{\"accessControlList\": [{\"grantee\": [{\"id\": \"*\"}],\"permission\": " +
				"[\"xxx\"]}]}",
			isSuc: false,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		var (
			aclWrapper *api.GetBucketAclResult
			aclOrg     *api.GetBucketAclResult
		)

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		err1 := testBosClient.PutBucketAclFromString(tCase.bucket, tCase.acl)
		if err1 == nil {
			aclWrapper, err = orgBosClient.GetBucketAcl(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketAcl II, ID %d : Get ACL failed", i+1)
				continue
			}
		}

		err2 := orgBosClient.PutBucketAclFromString(tCase.bucket, tCase.acl)
		if err2 == nil {
			aclOrg, err = orgBosClient.GetBucketAcl(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketAcl IV, ID %d : Get ACL failed: %s", i+1, err)
				continue
			}
		}

		util.ExpectEqual("wraper.go PutBucketAcl V", i+1, t.Errorf, tCase.isSuc,
			err1 == nil)
		if tCase.isSuc != (err1 == nil) {
			t.Errorf("PutBucketAcl suc, ID %d : want %v get %v, err: %s", i+1, tCase.isSuc,
				err1 == nil, err1)
			continue
		}

		if !tCase.isSuc {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go PutBucketAcl VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go PutBucketAcl VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go PutBucketAcl VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if aclWrapper == nil {
				t.Errorf("wraper.go PutBucketAcl IX: %d acl is nil", i+1)
			}

			util.ExpectEqual("wraper.go PutBucketAcl X", i+1, t.Errorf, true,
				aclWrapper != nil)
			util.ExpectEqual("wraper.go PutBucketAcl XI", i+1, t.Errorf, true,
				aclOrg != nil)

			util.ExpectEqual("wraper.go PutBucketAcl XII", i+1, t.Errorf, aclWrapper,
				aclOrg)
		}
	}
}

type getAclFromStringType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestGetBucketAclFromString(t *testing.T) {
	testCases := []getAclFromStringType{
		// 1
		getAclFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     true,
		},
		// 2
		getAclFromStringType{
			bucket:    bucket_gz,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			isSuc:     true,
		},
		// 3
		getAclFromStringType{
			bucket: bucket_bj,
			isSuc:  true,
		},

		// 5
		getAclFromStringType{
			bucket:    bucket_error,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			isSuc:     false,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		aclWrapper, err1 := testBosClient.GetBucketAcl(tCase.bucket)

		aclOrg, err2 := orgBosClient.GetBucketAcl(tCase.bucket)

		util.ExpectEqual("wraper.go GetBucketAcl V", i+1, t.Errorf, tCase.isSuc, err1 == nil)
		if tCase.isSuc != (err1 == nil) {
			t.Errorf("GetBucketAcl suc, ID %d : want %v get %v, err: %s", i+1, tCase.isSuc,
				err1 == nil, err1)
			continue
		}

		if !tCase.isSuc {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go GetBucketAcl VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go GetBucketAcl VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go GetBucketAcl VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if aclWrapper == nil {
				t.Errorf("wraper.go GetBucketAcl IX: %d acl is nil", i+1)
			}

			util.ExpectEqual("wraper.go GetBucketAcl X", i+1, t.Errorf, true,
				aclWrapper != nil)
			util.ExpectEqual("wraper.go GetBucketAcl XI", i+1, t.Errorf, true,
				aclOrg != nil)

			util.ExpectEqual("wraper.go GetBucketAcl XII", i+1, t.Errorf, aclWrapper,
				aclOrg)
		}
	}
}

// test Put Lifecycle from string
type putLifecycleFromStringType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	lifecycle string
	isSuc     bool
}

func TestPutBucketLifecycleFromString(t *testing.T) {
	var (
		err error
	)
	testCases := []putLifecycleFromStringType{
		// 1
		putLifecycleFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			lifecycle: "xxx",
			isSuc:     false,
		},
		// 2
		putLifecycleFromStringType{
			bucket:    bucket_bj,
			lifecycle: "xxx",
			isSuc:     false,
		},
		// 3
		putLifecycleFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_bj, bucket_bj, "dateGreaterThan"),
			isSuc:     true,
		},

		// 4
		putLifecycleFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_bj, bucket_bj, "dateGreaterThan"),
			isSuc:     true,
		},
		// 5
		putLifecycleFromStringType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_bj, bucket_bj, "xx"),
			isSuc:     false,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		var (
			lifecycleWrapper *api.GetBucketLifecycleResult
			lifecycleOrg     *api.GetBucketLifecycleResult
		)

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		err1 := testBosClient.PutBucketLifecycleFromString(tCase.bucket, tCase.lifecycle)
		if err1 == nil {
			lifecycleWrapper, err = orgBosClient.GetBucketLifecycle(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketLifecycle II, ID %d : Get lifecycle failed", i+1)
				continue
			}
		}

		err2 := orgBosClient.PutBucketLifecycleFromString(tCase.bucket, tCase.lifecycle)
		if err2 == nil {
			lifecycleOrg, err = orgBosClient.GetBucketLifecycle(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketLifecycle IV, ID %d : Get lifecycle failed: %s", i+1, err)
				continue
			}
		}

		util.ExpectEqual("wraper.go PutBucketLifecycle V", i+1, t.Errorf, tCase.isSuc,
			err1 == nil)
		if tCase.isSuc != (err1 == nil) {
			t.Logf(tCase.lifecycle)
			t.Errorf("PutBucketLifecycle suc, ID %d : want %v get %v, err: %s", i+1, tCase.isSuc,
				err1 == nil, err1)
			continue
		}

		if !tCase.isSuc {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go PutBucketLifecycle VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go PutBucketLifecycle VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go PutBucketLifecycle VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if lifecycleWrapper == nil {
				t.Errorf("wraper.go PutBucketLifecycle IX: %d lifecycle is nil", i+1)
			}

			util.ExpectEqual("wraper.go PutBucketLifecycle X", i+1, t.Errorf, true,
				lifecycleWrapper != nil)
			util.ExpectEqual("wraper.go PutBucketLifecycle XI", i+1, t.Errorf, true,
				lifecycleOrg != nil)

			util.ExpectEqual("wraper.go PutBucketLifecycle XII", i+1, t.Errorf, lifecycleWrapper,
				lifecycleOrg)
		}
	}
}

type getLifecycleType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestGetBucketLifecycle(t *testing.T) {
	testCases := []getLifecycleType{
		// 1
		getLifecycleType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
		},
		// 2
		getLifecycleType{
			bucket:    bucket_gz,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
		// 3
		getLifecycleType{
			bucket: bucket_bj,
		},

		// 5
		getLifecycleType{
			bucket:    bucket_error,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		// get lifecycle from wrapped bos client
		lifecycleWrapper, err1 := testBosClient.GetBucketLifecycle(tCase.bucket)
		// get lifecycle from original bos client
		lifecycleOrg, err2 := orgBosClient.GetBucketLifecycle(tCase.bucket)

		util.ExpectEqual("wraper.go GetBucketLifecycle V", i+1, t.Errorf, err1 == nil, err2 == nil)
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("GetBucketLifecycle suc, ID %d : err1 %s err2 %s", i+1, err1, err2)
			continue
		}

		if err1 != nil {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go GetBucketLifecycle VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go GetBucketLifecycle VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go GetBucketLifecycle VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if lifecycleWrapper == nil {
				t.Errorf("wraper.go GetBucketLifecycle IX: %d lifecycle is nil", i+1)
			}

			util.ExpectEqual("wraper.go GetBucketLifecycle X", i+1, t.Errorf, true,
				lifecycleWrapper != nil)
			util.ExpectEqual("wraper.go GetBucketLifecycle XI", i+1, t.Errorf, true,
				lifecycleOrg != nil)

			util.ExpectEqual("wraper.go GetBucketLifecycle XII", i+1, t.Errorf, lifecycleWrapper,
				lifecycleOrg)
		}
	}
}

// test Put Lifecycle from string
type deleteLifecycleType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	lifecycle string
}

func TestDeleteBucketLifecycle(t *testing.T) {

	testCases := []deleteLifecycleType{
		// 1
		deleteLifecycleType{
			bucket: bucket_gz,
		},
		// 2
		deleteLifecycleType{
			bucket:    bucket_gz,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_gz, bucket_gz, "dateGreaterThan"),
		},
		// 3
		deleteLifecycleType{
			bucket:    bucket_su,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_su, bucket_su, "dateGreaterThan"),
		},
		// 4
		deleteLifecycleType{
			bucket:    bucket_gz,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
			lifecycle: fmt.Sprintf(lifecycleTemp, bucket_gz, bucket_gz, "dateGreaterThan"),
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		// delete lifecycle
		orgBosClient.DeleteBucketLifecycle(tCase.bucket)

		// test delete lifecycle from orginal bos client
		if tCase.lifecycle != "" {
			orgBosClient.PutBucketLifecycleFromString(tCase.bucket, tCase.lifecycle)
		}
		err1 := testBosClient.DeleteBucketLifecycle(tCase.bucket)

		// test delete lifecycle from wrapped bos client
		if tCase.lifecycle != "" {
			orgBosClient.PutBucketLifecycleFromString(tCase.bucket, tCase.lifecycle)
		}
		err2 := orgBosClient.DeleteBucketLifecycle(tCase.bucket)

		// start compare
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("DeleteBucketLifecycle ID %d err1 %s err2 %s", i+1, err1, err2)
			continue
		}

		util.ExpectEqual("wraper.go deleteBucketLifecycle V", i+1, t.Errorf, err1 == nil,
			err2 == nil)

		if err1 != nil {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go deleteBucketLifecycle VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go deleteBucketLifecycle VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go deleteBucketLifecycle VIII", i+1, t.Errorf, err1, err2)
			}
		}
	}
}

// test Put logging
type putLoggingFromStructType struct {
	bucket       string
	cacheHave    bool
	endpoint     string
	TargetBucket string
	TargetPrefix string
	isSuc        bool
}

func TestPutBucketLoggingFromStruct(t *testing.T) {
	var (
		err error
	)

	testCases := []putLoggingFromStructType{
		// 1
		putLoggingFromStructType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			TargetBucket: bucket_gz,
			isSuc:        false,
		},
		// 2
		putLoggingFromStructType{
			bucket:       bucket_bj,
			TargetBucket: bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			isSuc:        false,
		},
		// 3
		putLoggingFromStructType{
			bucket:       bucket_bj,
			TargetBucket: bucket_bj1,
			TargetPrefix: "log",
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			isSuc:        true,
		},
		// 4
		putLoggingFromStructType{
			bucket:       bucket_bj,
			TargetBucket: bucket_error,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			isSuc:        false,
		},
		// 5
		putLoggingFromStructType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
			isSuc:     false,
		},
		// 6
		putLoggingFromStructType{
			bucket:       bucket_bj,
			TargetBucket: bucket_bj1,
			endpoint:     "gz.bcebos.com",
			isSuc:        false,
		},
		// 7
		putLoggingFromStructType{
			bucket:       bucket_error,
			TargetBucket: bucket_bj1,
			endpoint:     "bj.bcebos.com",
			isSuc:        false,
		},
		// 8
		putLoggingFromStructType{
			bucket:       bucket_bj,
			TargetBucket: bucket_gz,
			endpoint:     "bj.bcebos.com",
			isSuc:        false,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		var (
			loggingWrapper *api.GetBucketLoggingResult
		)

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		args := &api.PutBucketLoggingArgs{
			TargetBucket: tCase.TargetBucket,
			TargetPrefix: tCase.TargetPrefix,
		}

		// delete logging
		orgBosClient.DeleteBucketLogging(tCase.bucket)

		// put logging
		err1 := testBosClient.PutBucketLoggingFromStruct(tCase.bucket, args)
		if err1 != nil {
			util.ExpectEqual("wraper.go PutBucketLogging II", i+1, t.Errorf, tCase.isSuc,
				err1 == nil)
			continue
		}

		// get logging and verify logging
		loggingWrapper, err = orgBosClient.GetBucketLogging(tCase.bucket)
		if err != nil {
			t.Errorf("PutBucketLogging III, ID %d : Get logging failed", i+1)
			continue
		}

		util.ExpectEqual("wraper.go PutBucketLogging IV", i+1, t.Errorf, "enabled",
			loggingWrapper.Status)

		util.ExpectEqual("wraper.go PutBucketLogging V", i+1, t.Errorf, tCase.TargetBucket,
			loggingWrapper.TargetBucket)

		util.ExpectEqual("wraper.go PutBucketLogging VI", i+1, t.Errorf, tCase.TargetPrefix,
			loggingWrapper.TargetPrefix)

		// delete logging
		if err := orgBosClient.DeleteBucketLogging(tCase.bucket); err != nil {
			t.Errorf("PutBucketLogging VII, ID %d : Delete logging failed, error: %s", i+1, err)
			continue
		}
	}
}

type getLoggingType struct {
	bucket       string
	cacheHave    bool
	endpoint     string
	TargetBucket string
	TargetPrefix string
	status       string
}

func TestGetBucketLogging(t *testing.T) {
	testCases := []getLoggingType{
		// 1
		getLoggingType{
			bucket:    bucket_bj,
			cacheHave: true,
			status:    "disabled",
			endpoint:  "gz.bcebos.com",
		},
		// 2
		getLoggingType{
			bucket:       bucket_bj,
			status:       "enabled",
			TargetBucket: bucket_bj1,
			TargetPrefix: "log",
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
		},
		// 3
		getLoggingType{
			status: "disabled",
			bucket: bucket_bj,
		},
		// 4
		getLoggingType{
			bucket:       bucket_bj,
			status:       "enabled",
			TargetBucket: bucket_bj1,
			cacheHave:    true,
			endpoint:     "bj.bcebos.com",
		},
		// 5
		getLoggingType{
			bucket:    bucket_error,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("ID: %d, write %s %s to cache failed", i+1, tCase.bucket,
					tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("ID: %d, delete %s from caceh failed", i+1, tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		orgBosClient.DeleteBucketLogging(tCase.bucket)

		// put logging
		if tCase.status == "enabled" {
			args := &api.PutBucketLoggingArgs{
				TargetBucket: tCase.TargetBucket,
				TargetPrefix: tCase.TargetPrefix,
			}
			err := testBosClient.PutBucketLoggingFromStruct(tCase.bucket, args)
			if err != nil {
				t.Errorf("ID: %d, wraper.go GetBucketLogging put logging failed, error: %s", i+1,
					err)
				continue
			}
		}

		// get logging from wrapped bos client
		loggingWrapper, err1 := testBosClient.GetBucketLogging(tCase.bucket)
		// get logging from original bos client
		loggingOrg, err2 := orgBosClient.GetBucketLogging(tCase.bucket)

		util.ExpectEqual("wraper.go GetBucketLogging I", i+1, t.Errorf, err1 == nil, err2 == nil)
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("GetBucketLogging suc, ID %d : err1 %s err2 %s", i+1, err1, err2)
			continue
		}

		if err1 != nil {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go GetBucketLogging II", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go GetBucketLogging III", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go GetBucketLogging IV", i+1, t.Errorf, err1, err2)
			}
		} else {
			if loggingWrapper == nil {
				t.Errorf("wraper.go GetBucketLogging V: %d logging is nil", i+1)
			}

			util.ExpectEqual("wraper.go GetBucketLogging VI", i+1, t.Errorf, true,
				loggingWrapper != nil)
			util.ExpectEqual("wraper.go GetBucketLogging VII", i+1, t.Errorf, true,
				loggingOrg != nil)

			util.ExpectEqual("wraper.go GetBucketLogging VIII", i+1, t.Errorf, loggingWrapper,
				loggingOrg)
		}
	}
}

// test Put Logging from string
type deleteLoggingType struct {
	bucket       string
	cacheHave    bool
	endpoint     string
	TargetBucket string
	TargetPrefix string
	status       string
}

func TestDeleteBucketLogging(t *testing.T) {

	testCases := []deleteLoggingType{
		// 1
		deleteLoggingType{
			bucket: bucket_gz,
		},
		// 2
		deleteLoggingType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			status:       "enabled",
			TargetBucket: bucket_bj1,
			TargetPrefix: "log",
		},
		// 3
		deleteLoggingType{
			bucket:    bucket_su,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
		},
		// 4
		deleteLoggingType{
			bucket:    bucket_error,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		// delete logging
		orgBosClient.DeleteBucketLogging(tCase.bucket)

		// test delete logging from orginal bos client
		if tCase.status == "enabled" {
			args := &api.PutBucketLoggingArgs{
				TargetBucket: tCase.TargetBucket,
				TargetPrefix: tCase.TargetPrefix,
			}
			err := testBosClient.PutBucketLoggingFromStruct(tCase.bucket, args)
			if err != nil {
				t.Errorf("ID: %d, wraper.go DeleteBucketLogging put logging failed I, error: %s",
					i+1, err)
				continue
			}
		}
		err1 := testBosClient.DeleteBucketLogging(tCase.bucket)

		// test delete logging from wrapped bos client
		if tCase.status == "enabled" {
			args := &api.PutBucketLoggingArgs{
				TargetBucket: tCase.TargetBucket,
				TargetPrefix: tCase.TargetPrefix,
			}
			err := testBosClient.PutBucketLoggingFromStruct(tCase.bucket, args)
			if err != nil {
				t.Errorf("ID: %d, wraper.go DeleteBucketLogging put logging failed II, error: %s",
					i+1, err)
				continue
			}
		}
		err2 := orgBosClient.DeleteBucketLogging(tCase.bucket)

		// start compare
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("DeleteBucketLogging ID %d err1 %s err2 %s", i+1, err1, err2)
			continue
		}

		util.ExpectEqual("wraper.go deleteBucketLogging I", i+1, t.Errorf, err1 == nil, err2 == nil)

		if err1 != nil {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go deleteBucketLogging II", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go deleteBucketLogging III", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go deleteBucketLogging IV", i+1, t.Errorf, err1, err2)
			}
		}
	}
}

// test Put Storageclass from string
type putBucketStorageclassType struct {
	bucket       string
	cacheHave    bool
	endpoint     string
	storageclass string
	isSuc        bool
}

func TestPutBucketStorageclass(t *testing.T) {
	var (
		err error
	)
	testCases := []putBucketStorageclassType{
		// 1
		putBucketStorageclassType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "bj.bcebos.com",
			storageclass: "xxx",
			isSuc:        false,
		},
		// 2
		putBucketStorageclassType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			storageclass: "xxx",
			isSuc:        false,
		},
		// 3
		putBucketStorageclassType{
			bucket:       bucket_bj,
			storageclass: "xxx",
			isSuc:        false,
		},
		// 4
		putBucketStorageclassType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			storageclass: "COLD",
			isSuc:        true,
		},
		// 5
		putBucketStorageclassType{
			bucket:       bucket_bj,
			cacheHave:    true,
			endpoint:     "gz.bcebos.com",
			storageclass: "COLD",
			isSuc:        true,
		},
		// 6
		putBucketStorageclassType{
			bucket:       bucket_error,
			cacheHave:    true,
			endpoint:     "bj.bcebos.com",
			storageclass: "COLD",
			isSuc:        false,
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		var (
			storageclassWrapper string
			storageclassOrg     string
		)

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		err1 := testBosClient.PutBucketStorageclass(tCase.bucket, tCase.storageclass)
		if err1 == nil {
			storageclassWrapper, err = testBosClient.GetBucketStorageclass(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketStorageclass II, ID %d : Get storageclass failed", i+1)
				continue
			}
		}

		err2 := orgBosClient.PutBucketStorageclass(tCase.bucket, tCase.storageclass)
		if err2 == nil {
			storageclassOrg, err = orgBosClient.GetBucketStorageclass(tCase.bucket)
			if err != nil {
				t.Errorf("PutBucketStorageclass IV, ID %d : Get storageclass failed: %s", i+1, err)
				continue
			}
		}

		util.ExpectEqual("wraper.go PutBucketStorageclass V", i+1, t.Errorf, tCase.isSuc,
			err1 == nil)
		if tCase.isSuc != (err1 == nil) {
			t.Logf(tCase.storageclass)
			if err1 == nil {
				t.Errorf("PutBucketStorageclass suc, ID %d : want %v get %v", i+1, tCase.isSuc,
					err1 == nil)
			} else {
				t.Errorf("PutBucketStorageclass suc, ID %d : want %v get %v, err: %s", i+1,
					tCase.isSuc, err1 == nil, err1.Error())
			}
			continue
		}

		if !tCase.isSuc {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go PutBucketStorageclass VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go PutBucketStorageclass VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go PutBucketStorageclass VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if storageclassWrapper == "" {
				t.Errorf("wraper.go PutBucketStorageclass IX: %d storageclass is empty", i+1)
			}

			util.ExpectEqual("wraper.go PutBucketStorageclass X", i+1, t.Errorf, true,
				storageclassWrapper != "")
			util.ExpectEqual("wraper.go PutBucketStorageclass XI", i+1, t.Errorf, true,
				storageclassOrg != "")

			util.ExpectEqual("wraper.go PutBucketStorageclass XII", i+1, t.Errorf,
				storageclassWrapper, storageclassOrg)
		}
	}
}

type getStorageclassType struct {
	bucket    string
	cacheHave bool
	endpoint  string
	isSuc     bool
}

func TestGetBucketStorageclass(t *testing.T) {
	testCases := []getStorageclassType{
		// 1
		getStorageclassType{
			bucket:    bucket_bj,
			cacheHave: true,
			endpoint:  "gz.bcebos.com",
		},
		// 2
		getStorageclassType{
			bucket:    bucket_gz,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
		// 3
		getStorageclassType{
			bucket: bucket_bj,
		},

		// 5
		getStorageclassType{
			bucket:    bucket_error,
			cacheHave: true,
			endpoint:  "bj.bcebos.com",
		},
	}

	for i, tCase := range testCases {
		if tCase.cacheHave {
			if ok := bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint,
				1080); !ok {
				t.Errorf("write %s %s to cache failed", tCase.bucket, tCase.endpoint)
				continue
			}
		} else {
			err := bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
			if err != nil {
				t.Errorf("delete %s from caceh failed", tCase.bucket)
				continue
			}
		}

		if endpoint, ok := bucketDomains[tCase.bucket]; ok {
			orgBosClient.Config.Endpoint = endpoint
		}

		// get storageclass from wrapped bos client
		storageclassWrapper, err1 := testBosClient.GetBucketStorageclass(tCase.bucket)
		// get storageclass from original bos client
		storageclassOrg, err2 := orgBosClient.GetBucketStorageclass(tCase.bucket)

		util.ExpectEqual("wraper.go GetBucketStorageclass V", i+1, t.Errorf, err1 == nil,
			err2 == nil)
		if (err1 == nil) != (err2 == nil) {
			if err1 == nil {
				t.Errorf("GetBucketStorageclass suc, ID %d : err1 nil err2 %s", i+1, err2.Error())
			} else {
				t.Errorf("GetBucketStorageclass suc, ID %d : err1 %s err2 nil", i+1, err1.Error())
			}
			continue
		}

		if err1 != nil {
			if serverErr1, ok := err1.(*bce.BceServiceError); ok {
				serverErr2, ok := err2.(*bce.BceServiceError)
				util.ExpectEqual("wraper.go GetBucketStorageclass VI", i+1, t.Errorf, true, ok)
				if ok {
					util.ExpectEqual("wraper.go GetBucketStorageclass VII", i+1, t.Errorf,
						serverErr1.Code, serverErr2.Code)
				}
			} else {
				util.ExpectEqual("wraper.go GetBucketStorageclass VIII", i+1, t.Errorf, err1, err2)
			}
		} else {
			if storageclassWrapper == "" {
				t.Errorf("wraper.go GetBucketStorageclass IX: %d storageclass is empty", i+1)
			}

			util.ExpectEqual("wraper.go GetBucketStorageclass X", i+1, t.Errorf, true,
				storageclassWrapper != "")
			util.ExpectEqual("wraper.go GetBucketStorageclass XI", i+1, t.Errorf, true,
				storageclassOrg != "")

			util.ExpectEqual("wraper.go GetBucketStorageclass XII", i+1, t.Errorf,
				storageclassWrapper, storageclassOrg)
		}
	}
}
