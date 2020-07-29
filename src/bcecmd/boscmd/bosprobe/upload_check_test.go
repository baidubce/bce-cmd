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
	"fmt"
	"os"
	"testing"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"utils/util"
)

func init() {
	var (
		err error
	)
	if err := initConfig(); err != nil {
		os.Exit(1)
	}
	ak, ok := bceconf.CredentialProvider.GetAccessKey()
	if !ok {
		fmt.Println("get AK failed")
		os.Exit(1)
	}
	sk, ok := bceconf.CredentialProvider.GetSecretKey()
	if !ok {
		fmt.Println("get SK failed")
		os.Exit(1)
	}
	if bosClient, err = NewClient(ak, sk, ""); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type uploadRequestInitType struct {
	req    ProbeRequest
	exCode BosProbeErrorCode
}

// Test requestInit
func TestUploadRequestInit(t *testing.T) {
	requestCheckCases := []uploadRequestInitType{
		uploadRequestInitType{
			req:    nil,
			exCode: PROBE_INIT_REQUEST_FAILED,
		},
		uploadRequestInitType{
			req:    &UploadCheckRes{BucketName: "X"},
			exCode: CODE_SUCCESS,
		},
	}
	uCheck := &UploadCheck{}
	for i, uCase := range requestCheckCases {
		ret := uCheck.requestInit(uCase.req)
		util.ExpectEqual("uploadcheck", i, t.Errorf, uCase.exCode, ret)
	}
}

type uploadRequestCheckType struct {
	uCheck *UploadCheck
	exCode BosProbeErrorCode
}

// Test requestInit
func TestUploadRequestCheck(t *testing.T) {
	requestCheckCases := []uploadRequestCheckType{
		uploadRequestCheckType{
			uCheck: &UploadCheck{},
			exCode: LOCAL_ARGS_NO_BUCKET,
		},
		uploadRequestCheckType{
			uCheck: &UploadCheck{bucketName: "X"},
			exCode: CODE_SUCCESS,
		},
	}
	for i, uCase := range requestCheckCases {
		ret := uCase.uCheck.requestCheck()
		util.ExpectEqual("upload check requestCheck", i, t.Errorf, uCase.exCode, ret)
	}
}

type uploadGetEndpointType struct {
	uCheck           *UploadCheck
	expectedEndpoint string
	expectedCode     BosProbeErrorCode
}

// Test getEndPoint
func TestUploadGetEndpoint(t *testing.T) {
	getEndpointCases := []uploadGetEndpointType{
		uploadGetEndpointType{
			uCheck:           &UploadCheck{bucketName: "xxxx", endpoint: "su.bcebos.com"},
			expectedEndpoint: "su.bcebos.com",
			expectedCode:     CODE_SUCCESS,
		},
		uploadGetEndpointType{
			uCheck:       &UploadCheck{bucketName: "xxxx"},
			expectedCode: LOCAL_GET_ENDPOINT_BUCKET_FAILED,
		},
		uploadGetEndpointType{
			uCheck:       &UploadCheck{},
			expectedCode: LOCAL_GET_ENDPOINT_BUCKET_FAILED,
		},
		uploadGetEndpointType{
			uCheck:           &UploadCheck{bucketName: "liupeng-bj"},
			expectedEndpoint: "bj.bcebos.com",
			expectedCode:     CODE_SUCCESS,
		},
	}
	for i, uCase := range getEndpointCases {
		endpoint, ret, _ := uCase.uCheck.getEndpoint(bosClient)
		util.ExpectEqual("uploauCheck getEndpoint I", i, t.Errorf, uCase.expectedCode, ret)
		util.ExpectEqual("uploauCheck getEndpoint II", i, t.Errorf, uCase.expectedEndpoint,
			endpoint)
	}
}

type uploadFileType struct {
	bucketName   string
	objectName   string
	localPath    string
	kind         int //1 from file, 2 from string
	mod          int
	exObjectName string
	exResp       *serviceResp
	exErr        error
}

func TestUploadFile(t *testing.T) {
	uploadCases := []uploadFileType{
		uploadFileType{
			bucketName:   "liupeng-bj",
			kind:         1,
			localPath:    "./test_file/clod.rar",
			objectName:   "code",
			exObjectName: "code",
		},
		uploadFileType{
			bucketName: "liupeng-gz",
			kind:       2,
		},
		uploadFileType{
			kind:  3,
			exErr: fmt.Errorf("unsppourt model"),
		},
	}
	cCheck := &UploadCheck{}
	sdk := NewGoSdkFake(bosClient)
	for i, uCase := range uploadCases {
		if uCase.kind == 1 {
			ret, _ := cCheck.uploadFile(sdk, uCase.bucketName, uCase.objectName, uCase.localPath,
				"", 1)
			util.ExpectEqual("uploadCheck", i, t.Errorf, uCase.exObjectName, ret.debugId)
		} else if uCase.kind == 2 {
			fileSize := 10000
			content := util.GetRandomString(int64(fileSize))
			ret, _ := cCheck.uploadFile(sdk, uCase.bucketName, uCase.objectName, uCase.localPath,
				content, 1)
			util.ExpectEqual("uploadCheck", i, t.Errorf, fileSize, ret.status)
		} else {
			_, err := cCheck.uploadFile(sdk, uCase.bucketName, uCase.objectName, uCase.localPath,
				"", 2)
			util.ErrorEqual("dowloadCheck downloadFile V", i, t.Errorf, uCase.exErr, err)
		}
	}
}

type uploadImplType struct {
	uCheck       *UploadCheck
	exCode       BosProbeErrorCode
	exObjectName string
}

func TestUploadCheckImpl(t *testing.T) {
	// create files to test
	util.InitTestFiles()
	uploadImplCasesOnline := []uploadImplType{
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/clod.rar",
				objectName: "code",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "code",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/xcache.txt",
				objectName: "xcache.txt",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/xcache.txt",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
			},
			exCode: CODE_SUCCESS,
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "serverErr",
				localPath:  "./test_file/xcache.txt",
			},
			exCode:       boscmd.CODE_INVALID_BUCKET_NAME,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "clientErr",
				localPath:  "./test_file/xcache.txt",
				objectName: "test/",
			},
			exObjectName: "test/xcache.txt",
			exCode:       boscmd.CODE_INVALID_BUCKET_NAME,
		},
	}
	// 	sdk := NewGoSdkFake(bosClient)
	sdk := NewGoSdk(bosClient)
	for i, uCase := range uploadImplCasesOnline {
		_, deta, stat := uCase.uCheck.checkImpl(sdk)
		if uCase.exCode != CODE_SUCCESS {
			util.ExpectEqual("uploadCheck I", i, t.Errorf, uCase.exCode, stat.code)
		} else {
			if stat.code != CODE_SUCCESS {
				util.ExpectEqual("uploadCheck II", i, t.Errorf, uCase.exCode, stat.code)
				continue
			}
			if uCase.exObjectName != "" {
				util.ExpectEqual("uploadCheck III", i, t.Errorf, uCase.exObjectName,
					deta.objectName)
			}
		}
	}
	uploadImplCasesOffline := []uploadImplType{
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/clod.rar",
				objectName: "code",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "code",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/xcache.txt",
				objectName: "xcache.txt",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/xcache.txt",
			},
			exCode:       CODE_SUCCESS,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
			},
			exCode: CODE_SUCCESS,
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "serverErr",
				localPath:  "./test_file/xcache.txt",
			},
			exCode:       boscmd.CODE_INVALID_BUCKET_NAME,
			exObjectName: "xcache.txt",
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "liupeng-bj",
				localPath:  "./test_file/cache.txt",
			},
			exCode: boscmd.LOCAL_FILE_NOT_EXIST,
		},
		uploadImplType{
			uCheck: &UploadCheck{
				bucketName: "clientErr",
				localPath:  "./test_file/xcache.txt",
				objectName: "test/",
			},
			exObjectName: "test/xcache.txt",
			exCode:       boscmd.LOCAL_BCECLIENTERROR,
		},
	}
	sdk = NewGoSdkFake(bosClient)
	for i, uCase := range uploadImplCasesOffline {
		_, deta, stat := uCase.uCheck.checkImpl(sdk)
		if uCase.exCode != CODE_SUCCESS {
			util.ExpectEqual("uploadCheck I", i, t.Errorf, uCase.exCode, stat.code)
		} else {
			if stat.code != CODE_SUCCESS {
				util.ExpectEqual("uploadCheck II", i, t.Errorf, uCase.exCode, stat.code)
				continue
			}
			if uCase.exObjectName != "" {
				util.ExpectEqual("uploadCheck III", i, t.Errorf, uCase.exObjectName,
					deta.objectName)
			}
		}
	}
}
