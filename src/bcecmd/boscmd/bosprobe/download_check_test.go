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
	"path/filepath"
	"testing"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"utils/util"
)

var (
	dCheck *DownloadCheck
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

type downloadRequestInitType struct {
	req    ProbeRequest
	exCode BosProbeErrorCode
}

// Test requestInit
func TestDownloadRequestInit(t *testing.T) {
	requestCheckCases := []downloadRequestInitType{
		downloadRequestInitType{
			req:    nil,
			exCode: PROBE_INIT_REQUEST_FAILED,
		},
		downloadRequestInitType{
			req:    &DownloadCheckRes{BucketName: "X"},
			exCode: CODE_SUCCESS,
		},
	}
	dCheck := &DownloadCheck{}
	for i, dCase := range requestCheckCases {
		ret := dCheck.requestInit(dCase.req)
		util.ExpectEqual("DownloadCheck", i, t.Errorf, dCase.exCode, ret)
	}
}

type requstCheckTestType struct {
	dCheck   *DownloadCheck
	expected BosProbeErrorCode
}

// test function requestCheck of download_check
func TestRequestCheck(t *testing.T) {
	requstCheckCases := []requstCheckTestType{
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "123", bucketName: "bucket"},
			expected: LOCAL_ARGS_BOTH_URL_BUCKET_EXIST,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "123", objectName: "bucket"},
			expected: LOCAL_ARGS_BOTH_URL_OBJECT_EXIST,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "123", endpoint: "bucket"},
			expected: LOCAL_ARGS_BOTH_URL_ENDPOINT_EXIST,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{},
			expected: LOCAL_ARGS_NO_BUCKET_OR_URL,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "dsfsdf"},
			expected: CODE_SUCCESS,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "http://www.qq.com"},
			expected: CODE_SUCCESS,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{bucketName: "xxxx"},
			expected: CODE_SUCCESS,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{bucketName: "xxxx", endpoint: "su.bcebos.com"},
			expected: CODE_SUCCESS,
		},
		requstCheckTestType{
			dCheck: &DownloadCheck{fileUrl: "http://liupeng-bj.bj.bcebos.com/201711081529251020" +
				"6?authorization=bce-auth-v1%2Fab3e8280a5ff436eb5c5b9d7fa14fde9%2F2017-12-05T13" +
				"%3A43%3A48Z%2F300%2Fhost%2F520671e96980ce592baa0cc74cb87550322c03f87f3aab1770f" +
				"fab760fb39464"},
			expected: CODE_SUCCESS,
		},
		requstCheckTestType{
			dCheck:   &DownloadCheck{fileUrl: "h//www.qq.com"},
			expected: CODE_SUCCESS,
		},
	}
	for i, dcheckCase := range requstCheckCases {
		ret := dcheckCase.dCheck.requestCheck()
		util.ExpectEqual("downlaod check requestCheck", i, t.Errorf, dcheckCase.expected, ret)
	}
}

type getEndpointType struct {
	dCheck           *DownloadCheck
	expectedEndpoint string
	expectedCode     BosProbeErrorCode
}

// Test getEndPoint
func TestGetEndpoint(t *testing.T) {
	getEndpointCases := []getEndpointType{
		getEndpointType{
			dCheck:           &DownloadCheck{bucketName: "xxxx", endpoint: "su.bcebos.com"},
			expectedEndpoint: "su.bcebos.com",
			expectedCode:     CODE_SUCCESS,
		},
		getEndpointType{
			dCheck:       &DownloadCheck{fileUrl: "xxxx"},
			expectedCode: LOCAL_PROBE_URL_IS_INVALID,
		},
		getEndpointType{
			dCheck: &DownloadCheck{fileUrl: "http://liupeng-bj.bj.bcebos.com/201711081529251020" +
				"6?authorization=bce-auth-v1%2Fab3e8280a5ff436eb5c5b9d7fa14fde9%2F2017-12-05T13" +
				"%3A43%3A48Z%2F300%2Fhost%2F520671e96980ce592baa0cc74cb87550322c03f87f3aab1770f" +
				"fab760fb39464"},
			expectedEndpoint: "liupeng-bj.bj.bcebos.com",
			expectedCode:     CODE_SUCCESS,
		},
		getEndpointType{
			dCheck:       &DownloadCheck{fileUrl: "//xxxx"},
			expectedCode: LOCAL_PROBE_URL_IS_INVALID,
		},
		getEndpointType{
			dCheck:       &DownloadCheck{bucketName: "xxxx"},
			expectedCode: LOCAL_GET_ENDPOINT_BUCKET_FAILED,
		},
		getEndpointType{
			dCheck:       &DownloadCheck{},
			expectedCode: LOCAL_ARGS_NO_BUCKET_OR_URL,
		},
		getEndpointType{
			dCheck:           &DownloadCheck{bucketName: "liupeng-bj"},
			expectedEndpoint: "bj.bcebos.com",
			expectedCode:     CODE_SUCCESS,
		},
	}
	for i, gEndCase := range getEndpointCases {
		endpoint, ret, _ := gEndCase.dCheck.getEndpoint(bosClient)
		util.ExpectEqual("dowloadCheck getEndpoint I", i, t.Errorf, gEndCase.expectedCode, ret)
		util.ExpectEqual("dowloadCheck getEndpoint II", i, t.Errorf, gEndCase.expectedEndpoint,
			endpoint)
	}
}

type downloadFileType struct {
	fileUrl     string
	bucketName  string
	objectName  string
	localPath   string
	exResp      *serviceResp
	exLocalPath string
	exErr       error
}

func TestDownloadFile(t *testing.T) {
	downloadCases := []downloadFileType{
		downloadFileType{
			fileUrl: "http://bj.bcebos.com/liupeng-bj/dexgdf?authorization=bce-auth-v1%2F637986f4" +
				"1b0046248e3a333817371502%2F2017-12-06T05%3A35%3A36Z%2F-1%2F%2F4bc7c59ce323ce72a4" +
				"5be1bff4ba17842bd15b5d6dddabf6cc7155a42d72deb8",
			localPath: generateRandomFileName(),
			exErr: &bce.BceServiceError{
				Code: "NoSuchKey",
			},
		},
		downloadFileType{
			fileUrl: "http://bj.bcebos.com/liupeng-bj/test.txt?authorization=bce-auth-v1%2F637986" +
				"f41b0046248e3a333817371502%2F2017-12-06T06%3A21%3A31Z%2F-1%2F%2F8993c513ea868d9e" +
				"711765d240956653c2bdad273e2c3bdab5f735cdef83ea61",
			exLocalPath: "test.txt",
		},
		downloadFileType{
			fileUrl: "http://bj.bcebos.com/liupeng-bj/test.txt?authorization=bce-auth-v1%2F637986" +
				"f41b0046248e3a333817371502%2F2017-12-06T06%3A21%3A31Z%2F-1%2F%2F8993c513ea868d9e" +
				"711765d240956653c2bdad273e2c3bdab5f735cdef83ea61",
			localPath:   "test1.txt",
			exLocalPath: "test1.txt",
		},
		downloadFileType{
			bucketName: "liupeng-bj",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			objectName:  "test.txt",
			localPath:   "./test.txt",
			exLocalPath: "./test.txt",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			localPath:   "./xxxx/xxx/test.txt",
			exLocalPath: "./xxxx/xxx/test.txt",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			objectName:  "test.txt",
			localPath:   "./xxxx/",
			exLocalPath: "./xxxx/test.txt",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			objectName:  "test.txt",
			localPath:   "./xxxx",
			exLocalPath: "./xxxx/test.txt",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			objectName:  "test.txt",
			localPath:   "./xxx/",
			exLocalPath: "./xxx/test.txt",
		},
		downloadFileType{
			bucketName:  "liupeng-bj",
			objectName:  "test.txt",
			localPath:   "./xxxxx/xxx/",
			exLocalPath: "./xxxxx/xxx/test.txt",
		},
	}
	dCheck := &DownloadCheck{}
	sdk := NewGoSdk(bosClient)
	for i, downCase := range downloadCases {
		_, localPath, err := dCheck.downloadFile(sdk, downCase.fileUrl, downCase.bucketName,
			downCase.objectName, downCase.localPath)
		util.ExpectEqual("dowloadCheck downloadFile ", i, t.Errorf, downCase.exErr == nil,
			err == nil)
		if downCase.exErr == nil && err != nil {
			t.Errorf("%s id: %d expect error is nil but get %s\n", "dowloadCheck download"+
				"File I", i, err)
			continue
		}
		if downCase.exErr == nil {
			if err != nil {
				t.Errorf("%s id: %d expect error is nil but get %s\n", "dowloadCheck download"+
					"File I", i, err)
				continue
			}

			if downCase.exLocalPath == "" {
				if !util.DoesFileExist(localPath) {
					t.Errorf("%s id: %d file %s don't exist\n", "dowloadCheck downloadFile II", i,
						localPath)
				}
			} else if !util.DoesFileExist(downCase.exLocalPath) {
				t.Errorf("%s id: %d ex file %s don't exist\n", "dowloadCheck downloadFile III", i,
					downCase.exLocalPath)
			} else {
				exAbsPath, _ := filepath.Abs(downCase.exLocalPath)
				absPath, _ := filepath.Abs(localPath)
				util.ExpectEqual("dowloadCheck downloadFile IV", i, t.Errorf, exAbsPath, absPath)
			}
			os.Remove(localPath)
		} else {
			util.ErrorEqual("dowloadCheck downloadFile V", i, t.Errorf, downCase.exErr, err)
		}
	}
}

type checkImplType struct {
	dCheck       *DownloadCheck
	expectedCode BosProbeErrorCode
	exLocalPath  string
}

func TestCheckImpl(t *testing.T) {
	sdk := NewGoSdk(bosClient)
	checkImplCases := []checkImplType{
		checkImplType{
			dCheck:       &DownloadCheck{bucketName: "dsfdsafsaf", endpoint: "su.bcebos.com"},
			expectedCode: boscmd.CODE_NO_SUCH_BUCKET,
		},
		checkImplType{
			dCheck:       &DownloadCheck{fileUrl: "xxxx"},
			expectedCode: boscmd.LOCAL_BCECLIENTERROR,
		},
		checkImplType{
			dCheck: &DownloadCheck{fileUrl: "http://bj.bcebos.com/liupeng-bj/test.txt?authoriz" +
				"ation=bce-auth-v1%2F637986" +
				"f41b0046248e3a333817371502%2F2017-12-06T06%3A21%3A31Z%2F-1%2F%2F8993c513ea868d9e" +
				"711765d240956653c2bdad273e2c3bdab5f735cdef83ea61"},
			expectedCode: CODE_SUCCESS,
			exLocalPath:  "test.txt",
		},
		checkImplType{
			dCheck:       &DownloadCheck{fileUrl: "//xxxx"},
			expectedCode: boscmd.LOCAL_BCECLIENTERROR,
		},
		checkImplType{
			dCheck:       &DownloadCheck{bucketName: "liupeng-bj"},
			expectedCode: CODE_SUCCESS,
		},
		checkImplType{
			dCheck: &DownloadCheck{
				bucketName: "liupeng-bj",
				objectName: "test.txt",
				localPath:  "/xxxxx/xxx/",
			},
			expectedCode: boscmd.LOCAL_BCECLIENTERROR,
		},
	}
	for i, tCase := range checkImplCases {
		_, deta, stat := tCase.dCheck.checkImpl(sdk)
		if tCase.expectedCode != stat.code {
			util.ExpectEqual("downlaod checkImp I", i, t.Errorf, tCase.expectedCode, stat.code)
			continue
		} else if tCase.expectedCode == CODE_SUCCESS {
			if tCase.exLocalPath != "" {
				util.ExpectEqual("downlaod checkImp II", i, t.Errorf, tCase.exLocalPath,
					deta.objectName)
			}
		}
	}
}
