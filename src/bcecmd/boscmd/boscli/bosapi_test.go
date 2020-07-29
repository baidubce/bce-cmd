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
	"io/ioutil"
	// 	"net/http"
	"os"
	// 	"runtime"
	// 	"path/filepath"
	// 	"strconv"
	// 	"strings"
	"testing"
)

import (
	"bcecmd/boscmd"
	// 	"bceconf"
	// 	"github.com/baidubce/bce-sdk-go/services/bos"
	// 	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"utils/util"
)

var (
	bosapi *BosApi
)

func init() {
	bosapi = NewBosApi()
	bosapi.bosClient = &fakeBosClientForBos{}
}

// Fake of PutBucketLifecycle
func (b *fakeBosClientForBos) PutBucketLifecycleFromString(bucket, lifecycle string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket + lifecycle)
}

// Fake of GetBucketLifecycle
func (b *fakeBosClientForBos) GetBucketLifecycle(bucket string) (*api.GetBucketLifecycleResult,
	error) {
	if bucket == "success" {
		return &api.GetBucketLifecycleResult{
			Rule: []api.LifecycleRuleType{
				api.LifecycleRuleType{
					Id:     "123",
					Status: "enabled",
				},
			},
		}, nil
	}
	return nil, fmt.Errorf(bucket)
}

// Fake of DeleteBucketLifecycle
func (b *fakeBosClientForBos) DeleteBucketLifecycle(bucket string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket)
}

// Fake of PutBucketLoggingFromStruct
func (b *fakeBosClientForBos) PutBucketLoggingFromStruct(bucket string,
	obj *api.PutBucketLoggingArgs) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket + obj.TargetBucket + obj.TargetPrefix)
}

// Fake of GetBucketLogging
func (b *fakeBosClientForBos) GetBucketLogging(bucket string) (*api.GetBucketLoggingResult, error) {
	if bucket == "success" {
		return &api.GetBucketLoggingResult{
			Status:       "enabled",
			TargetBucket: "success1",
			TargetPrefix: "log1",
		}, nil
	}
	return nil, fmt.Errorf(bucket)
}

// Fake of DeleteBucketLogging
func (b *fakeBosClientForBos) DeleteBucketLogging(bucket string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket)
}

// Fake of PutBucketStorageclass
func (b *fakeBosClientForBos) PutBucketStorageclass(bucket, storageClass string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket + storageClass)
}

// Fake of GetBucketStorageclass
func (b *fakeBosClientForBos) GetBucketStorageclass(bucket string) (string, error) {
	if bucket == "success" {
		return "COLD", nil
	}
	return "", fmt.Errorf(bucket)
}

// Fake of PutBucketAclFromCanned
func (b *fakeBosClientForBos) PutBucketAclFromCanned(bucket, cannedAcl string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket + cannedAcl)
}

// Fake of PutBucketAcl
func (b *fakeBosClientForBos) PutBucketAclFromString(bucket, acl string) error {
	if bucket == "success" {
		return nil
	}
	return fmt.Errorf(bucket + acl)
}

// Fake of GetBucketAcl
func (b *fakeBosClientForBos) GetBucketAcl(bucket string) (*api.GetBucketAclResult, error) {
	if bucket == "success" {
		return &api.GetBucketAclResult{
			AccessControlList: []api.GrantType{},
			Owner: api.AclOwnerType{
				Id: "123",
			},
		}, nil
	}
	return nil, fmt.Errorf(bucket)
}

type putBucketAclPreProcessType struct {
	configPath string
	bosPath    string
	canned     string
	bucketName string
	acl        []byte
	opType     int
	code       BosCliErrorCode
}

func TestPutBucketAclPreProcess(t *testing.T) {
	fd, fileName, err := util.CreateAnRandomFileWithContent("%s",
		`{"accessControlList": [{"grantee": [{"id": "*"}],"permission": ["READ"]}]}`)
	if err != nil {
		t.Errorf("create acl test file failed! error: %v", err)
		return
	}

	acl, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Errorf("get acl from file failed! error: %v", err)
		return
	}

	fd.Close()
	defer os.Remove(fileName)

	testCases := []putBucketAclPreProcessType{
		// 1
		putBucketAclPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		putBucketAclPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		putBucketAclPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucketkey",
			configPath: "./test",
			canned:     "private",
			code:       BOSCLI_PUT_ACL_CANNED_FILE_SAME_TIME,
		},
		// 5
		putBucketAclPreProcessType{
			bosPath: "bos:/bucketkey",
			canned:  "xxprivate",
			code:    BOSCLI_PUT_ACL_CANNED_DONT_SUPPORT,
		},
		// 6
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			canned:     "private",
			bucketName: "bucket",
			opType:     2,
			code:       BOSCLI_OK,
		},
		// 7
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			configPath: "./acltest",
			code:       boscmd.LOCAL_FILE_NOT_EXIST,
		},
		// 8
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			configPath: "./bosapi.go",
			code:       BOSCLI_EMPTY_CODE,
		},
		// 9
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			configPath: "./bosapi.go",
			code:       BOSCLI_EMPTY_CODE,
		},
		// 10
		putBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			configPath: fileName,
			code:       BOSCLI_OK,
			opType:     1,
			bucketName: "bucket",
			acl:        acl,
		},
		// 11
		putBucketAclPreProcessType{
			bosPath: "bos:/bucket",
			code:    BOSCLI_PUT_ACL_CANNED_FILE_BOTH_EMPTY,
		},
		// 12
		putBucketAclPreProcessType{
			bosPath:    "bos://bucket",
			canned:     "private",
			bucketName: "bucket",
			opType:     2,
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, _, code := bosapi.putBucketAclPreProcess(tCase.configPath, tCase.bosPath, tCase.canned)
		util.ExpectEqual("bosapi.go putBucketAclPreProcess I", i+1, t.Errorf, tCase.code, code)
		if code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go putBucketAclPreProcess II", i+1, t.Errorf, tCase.bucketName,
				ret.bucketName)
			util.ExpectEqual("bosapi.go putBucketAclPreProcess III", i+1, t.Errorf, tCase.opType,
				ret.opType)
			util.ExpectEqual("bosapi.go putBucketAclPreProcess IV", i+1, t.Errorf, tCase.acl,
				ret.acl)
		}
	}
}

type putBucketAclType struct {
	configPath string
	bosPath    string
	canned     string
}

func TestPutBucketAcl(t *testing.T) {
	fd, fileName, err := util.CreateAnRandomFileWithContent("%s",
		`{"accessControlList": [{"grantee": [{"id": "*"}],"permission": ["READ"]}]}`)
	if err != nil {
		t.Errorf("create acl test file failed! error: %v", err)
		return
	}

	fd.Close()
	defer os.Remove(fileName)

	testCases := []putBucketAclType{
		// 1
		putBucketAclType{
			bosPath: "bos:/success",
			canned:  "private",
		},
		// 10
		putBucketAclType{
			bosPath:    "bos:/success",
			configPath: fileName,
		},
	}
	for _, tCase := range testCases {
		bosapi.PutBucketAcl(tCase.configPath, tCase.bosPath, tCase.canned)
	}
}

type putBucketAclExecuteType struct {
	opType     int
	aclJosn    []byte
	bucketName string
	canned     string
	code       BosCliErrorCode
	err        string
}

func TestPutBucketAclExecute(t *testing.T) {
	acl := `{"accessControlList": [{"grantee": [{"id": "*"}],"permission": ["READ"]}]}`

	testCases := []putBucketAclExecuteType{
		// 1
		putBucketAclExecuteType{
			opType:     1,
			bucketName: "success",
			aclJosn:    []byte(acl),
			code:       BOSCLI_OK,
		},
		// 2
		putBucketAclExecuteType{
			opType:     1,
			bucketName: "error",
			aclJosn:    []byte(acl),
			err:        "error" + acl,
			code:       BOSCLI_EMPTY_CODE,
		},
		// 3
		putBucketAclExecuteType{
			opType:     2,
			bucketName: "error",
			canned:     "private",
			err:        "errorprivate",
			code:       BOSCLI_EMPTY_CODE,
		},
		// 4
		putBucketAclExecuteType{
			opType:     2,
			bucketName: "success",
			canned:     "private",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		err, code := bosapi.putBucketAclExecute(tCase.opType, tCase.aclJosn, tCase.bucketName,
			tCase.canned)
		util.ExpectEqual("bosapi.go putBucketAclExecute I", i+1, t.Errorf, tCase.code, code)
		if code != BOSCLI_OK {
			util.ExpectEqual("bosapi.go putBucketAclExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

type getBucketAclPreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestGetBucketAclPreProcess(t *testing.T) {
	testCases := []getBucketAclPreProcessType{
		// 1
		getBucketAclPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		getBucketAclPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		getBucketAclPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		getBucketAclPreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.getBucketAclPreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go getBucketAclPreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go getBucketAclPreProcess II", i+1, t.Errorf, tCase.bucketName,
				ret)
		}
	}
}

type getBucketAclExecuteType struct {
	bucketName string
	ownerId    string
	err        string
}

func TestGetBucketAclExecute(t *testing.T) {
	testCases := []getBucketAclExecuteType{
		// 1
		getBucketAclExecuteType{
			bucketName: "success",
			ownerId:    "123",
		},
		// 2
		getBucketAclExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		getBucketAclExecuteType{
			bucketName: "error1",
			err:        "error1",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.getBucketAclExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go getBucketAclExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go getBucketAclExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestGetBucketAcl(t *testing.T) {
	bosapi.GetBucketAcl("success")
}

var (
	lifecycle = `
		{
			"rule": [
			{
				 "status": "enabled",
		 		 "action": {
		 				"name": "Transition",
		 				"storageClass": "STANDARD_IA"
		 			},
		 			"resource": [
		 					"liupeng-bj/*"
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
		 					"liupeng-bj/*"
		 			],
		 			"id": "sample-rule-transition-2",
		 			"condition": {
		 				"time": {
		 					"dateGreaterThan": "$(lastModified)+P300D"
						}
					},
					"index":"index.html"
			  }
		
			]
		}
	`
)

type putLifecyclePreProcessType struct {
	configPath string
	bosPath    string
	template   bool
	bucketName string
	lifecycle  []byte
	code       BosCliErrorCode
}

func TestPutLifecyclePreProcess(t *testing.T) {

	fd, fileName, err := util.CreateAnRandomFileWithContent("%s", lifecycle)
	if err != nil {
		t.Errorf("create lifecycle test file failed! error: %v", err)
		return
	}

	lifecycleJosn, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Errorf("get lifecycle from file failed! error: %v", err)
		return
	}

	fd.Close()
	defer os.Remove(fileName)

	testCases := []putLifecyclePreProcessType{
		// 1
		putLifecyclePreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		putLifecyclePreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		putLifecyclePreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		putLifecyclePreProcessType{
			bosPath:    "bos:/bucket",
			configPath: "./lifecycletest",
			code:       boscmd.LOCAL_FILE_NOT_EXIST,
		},
		// 5
		putLifecyclePreProcessType{
			bosPath:    "bos:/bucket",
			configPath: "./bosapi.go",
			code:       BOSCLI_EMPTY_CODE,
		},
		// 6
		putLifecyclePreProcessType{
			bosPath:    "bos:/bucket",
			configPath: fileName,
			code:       BOSCLI_OK,
			bucketName: "bucket",
			lifecycle:  lifecycleJosn,
		},
		// 7
		putLifecyclePreProcessType{
			bosPath: "bos:/bucket",
			code:    BOSCLI_PUT_LIFECYCLE_NO_CONFIG_AND_BUCKET,
		},
	}

	for i, tCase := range testCases {
		ret, _, code := bosapi.putLifecyclePreProcess(tCase.configPath, tCase.bosPath,
			tCase.template)
		util.ExpectEqual("bosapi.go putLifecyclePreProcess I", i+1, t.Errorf, tCase.code, code)
		if code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go putLifecyclePreProcess II", i+1, t.Errorf, tCase.bucketName,
				ret.bucketName)
			util.ExpectEqual("bosapi.go putLifecyclePreProcess IV", i+1, t.Errorf, tCase.lifecycle,
				ret.lifecycle)
		}
	}
}

type putLifecycleType struct {
	configPath string
	bosPath    string
	template   bool
}

func TestPutLifecycle(t *testing.T) {
	fd, fileName, err := util.CreateAnRandomFileWithContent("%s", lifecycle)
	if err != nil {
		t.Errorf("create lifecycle test file failed! error: %v", err)
		return
	}

	fd.Close()
	defer os.Remove(fileName)

	testCases := []putLifecycleType{
		// 1
		putLifecycleType{
			bosPath:    "bos:/success",
			configPath: fileName,
		},
		// 2
		putLifecycleType{
			bosPath:    "bos:/success",
			configPath: fileName,
			template:   true,
		},
		// 3
		putLifecycleType{
			bosPath:  "bos:/success",
			template: true,
		},
	}

	for i, tCase := range testCases {
		fmt.Println("\nstart:", i+1)
		bosapi.PutLifecycle(tCase.configPath, tCase.bosPath, tCase.template)
	}
}

type putLifecycleExecuteType struct {
	lifecycleJosn []byte
	bucketName    string
	template      bool
	code          BosCliErrorCode
	err           string
}

func TestPutLifecycleExecute(t *testing.T) {
	testCases := []putLifecycleExecuteType{
		// 1
		putLifecycleExecuteType{
			bucketName:    "success",
			lifecycleJosn: []byte(lifecycle),
			code:          BOSCLI_OK,
		},
		// 2
		putLifecycleExecuteType{
			template: true,
			code:     BOSCLI_OK,
		},
		// 3
		putLifecycleExecuteType{
			bucketName:    "error",
			lifecycleJosn: []byte(lifecycle),
			err:           "error" + lifecycle,
			code:          BOSCLI_EMPTY_CODE,
		},
	}

	for i, tCase := range testCases {
		err, code := bosapi.putLifecycleExecute(tCase.lifecycleJosn, tCase.bucketName,
			tCase.template)
		util.ExpectEqual("bosapi.go putLifecycleExecute I", i+1, t.Errorf, tCase.code, code)
		if code != BOSCLI_OK {
			util.ExpectEqual("bosapi.go putLifecycleExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

type getLifecyclePreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestGetLifecyclePreProcess(t *testing.T) {
	testCases := []getLifecyclePreProcessType{
		// 1
		getLifecyclePreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		getLifecyclePreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		getLifecyclePreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		getLifecyclePreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.getLifecyclePreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go getLifecyclePreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go getLifecyclePreProcess II", i+1, t.Errorf, tCase.bucketName,
				ret)
		}
	}
}

type getLifecycleExecuteType struct {
	bucketName string
	err        string
}

func TestGetLifecycleExecute(t *testing.T) {
	testCases := []getLifecycleExecuteType{
		// 1
		getLifecycleExecuteType{
			bucketName: "success",
		},
		// 2
		getLifecycleExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		getLifecycleExecuteType{
			bucketName: "error1",
			err:        "error1",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.getLifecycleExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go getLifecycleExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go getLifecycleExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestGetLifecycle(t *testing.T) {
	bosapi.GetLifecycle("success")
}

type deleteLifecyclePreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestDeleteLifecyclePreProcess(t *testing.T) {
	testCases := []deleteLifecyclePreProcessType{
		// 1
		deleteLifecyclePreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		deleteLifecyclePreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		deleteLifecyclePreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		deleteLifecyclePreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.deleteLifecyclePreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go deleteLifecyclePreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go deleteLifecyclePreProcess II", i+1, t.Errorf,
				tCase.bucketName, ret)
		}
	}
}

type deleteLifecycleExecuteType struct {
	bucketName string
	err        string
}

func TestDeleteLifecycleExecute(t *testing.T) {
	testCases := []deleteLifecycleExecuteType{
		// 1
		deleteLifecycleExecuteType{
			bucketName: "success",
		},
		// 2
		deleteLifecycleExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		deleteLifecycleExecuteType{
			bucketName: "error1",
			err:        "error1",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.deleteLifecycleExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go deleteLifecycleExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go deleteLifecycleExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestDeleteLifecycle(t *testing.T) {
	bosapi.DeleteLifecycle("success")
}

type putLoggingPreProcessType struct {
	targetBosPath string
	targetPrefix  string
	bosPath       string
	targetName    string
	bucketName    string
	code          BosCliErrorCode
}

func TestPutLoggingPreProcess(t *testing.T) {

	testCases := []putLoggingPreProcessType{
		// 1
		putLoggingPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		putLoggingPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		putLoggingPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		putLoggingPreProcessType{
			bosPath:       "bos:/bucket",
			targetBosPath: "/bos",
			code:          BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		putLoggingPreProcessType{
			bosPath:       "bos:/bucket",
			targetBosPath: "liup/object",
			code:          BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		putLoggingPreProcessType{
			bosPath:       "bos:/bucket",
			targetBosPath: "bos://",
			code:          BOSCLI_PUT_LOG_NO_TARGET_BUCKET,
		},
		// 5
		putLoggingPreProcessType{
			bosPath:       "bos:/bucket",
			targetBosPath: "bos:/bucket1/dsf",
			code:          BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 6
		putLoggingPreProcessType{
			bosPath:       "bos:/bucket",
			targetBosPath: "bos:/bucket1/",
			code:          BOSCLI_OK,
			bucketName:    "bucket",
			targetName:    "bucket1",
		},
	}

	for i, tCase := range testCases {
		bucketName, targetName, code := bosapi.putLoggingPreProcess(tCase.targetBosPath,
			tCase.targetPrefix, tCase.bosPath)
		util.ExpectEqual("bosapi.go putLoggingPreProcess I", i+1, t.Errorf, tCase.code, code)
		if code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go putLoggingPreProcess II", i+1, t.Errorf, tCase.bucketName,
				bucketName)
			util.ExpectEqual("bosapi.go putLoggingPreProcess IV", i+1, t.Errorf, tCase.targetName,
				targetName)
		}
	}
}

type putLoggingType struct {
	targetBosPath string
	targetPrefix  string
	bosPath       string
}

func TestPutLogging(t *testing.T) {

	testCases := []putLoggingType{
		// 1
		putLoggingType{
			targetBosPath: "bos:/success",
			bosPath:       "bos:/success",
		},
		// 2
		putLoggingType{
			targetBosPath: "bos:/success",
			bosPath:       "bos:/success",
			targetPrefix:  "log",
		},
	}

	for i, tCase := range testCases {
		fmt.Println("\nstart:", i+1)
		bosapi.PutLogging(tCase.targetBosPath, tCase.targetPrefix, tCase.bosPath)
	}
}

type putLoggingExecuteType struct {
	targetName   string
	targetPrefix string
	bucketName   string
	err          string
}

func TestPutLoggingExecute(t *testing.T) {
	testCases := []putLoggingExecuteType{
		// 1
		putLoggingExecuteType{
			bucketName:   "success",
			targetName:   "success",
			targetPrefix: "log",
		},
		// 2
		putLoggingExecuteType{
			bucketName: "success",
			targetName: "success",
		},
		// 3
		putLoggingExecuteType{
			bucketName:   "error",
			targetName:   "success",
			targetPrefix: "log",
			err:          "errorsuccesslog",
		},
	}

	for i, tCase := range testCases {
		err := bosapi.putLoggingExecute(tCase.targetName, tCase.targetPrefix, tCase.bucketName)
		util.ExpectEqual("bosapi.go putLoggingExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if tCase.err != "" {
			util.ExpectEqual("bosapi.go putLoggingExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

type getLoggingPreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestGetLoggingPreProcess(t *testing.T) {
	testCases := []getLoggingPreProcessType{
		// 1
		getLoggingPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		getLoggingPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		getLoggingPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		getLoggingPreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.getLoggingPreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go getLoggingPreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go getLoggingPreProcess II", i+1, t.Errorf, tCase.bucketName,
				ret)
		}
	}
}

type getLoggingExecuteType struct {
	bucketName string
	err        string
}

func TestGetLoggingExecute(t *testing.T) {
	testCases := []getLoggingExecuteType{
		// 1
		getLoggingExecuteType{
			bucketName: "success",
		},
		// 2
		getLoggingExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		getLoggingExecuteType{
			bucketName: "errorlogging",
			err:        "errorlogging",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.getLoggingExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go getLoggingExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go getLoggingExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestGetLogging(t *testing.T) {
	bosapi.GetLogging("success")
}

type deleteLoggingPreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestDeleteLoggingPreProcess(t *testing.T) {
	testCases := []deleteLoggingPreProcessType{
		// 1
		deleteLoggingPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		deleteLoggingPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		deleteLoggingPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		deleteLoggingPreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.deleteLoggingPreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go deleteLoggingPreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go deleteLoggingPreProcess II", i+1, t.Errorf,
				tCase.bucketName, ret)
		}
	}
}

type deleteLoggingExecuteType struct {
	bucketName string
	err        string
}

func TestDeleteLoggingExecute(t *testing.T) {
	testCases := []deleteLoggingExecuteType{
		// 1
		deleteLoggingExecuteType{
			bucketName: "success",
		},
		// 2
		deleteLoggingExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		deleteLoggingExecuteType{
			bucketName: "errorlogging",
			err:        "errorlogging",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.deleteLoggingExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go deleteLoggingExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go deleteLoggingExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestDeleteLogging(t *testing.T) {
	bosapi.DeleteLogging("success")
}

type putBucketStorageClassPreProcessType struct {
	bosPath      string
	storageClass string
	bucketName   string
	code         BosCliErrorCode
}

func TestPutBucketStorageClassPreProcess(t *testing.T) {

	testCases := []putBucketStorageClassPreProcessType{
		// 1
		putBucketStorageClassPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		putBucketStorageClassPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		putBucketStorageClassPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		putBucketStorageClassPreProcessType{
			bosPath: "bos:/bucket",
			code:    BOSCLI_STORAGE_CLASS_IS_EMPTY,
		},
		// 5
		putBucketStorageClassPreProcessType{
			bosPath:      "bos:/bucket",
			storageClass: "xxx",
			code:         BOSCLI_UNSUPPORT_STORAGE_CLASS,
		},
		// 6
		putBucketStorageClassPreProcessType{
			bosPath:      "bos:/bucket",
			storageClass: "COLD",
			bucketName:   "bucket",
			code:         BOSCLI_OK,
		},
	}

	for i, tCase := range testCases {
		bucketName, code := bosapi.putBucketStorageClassPreProcess(tCase.bosPath,
			tCase.storageClass)
		util.ExpectEqual("bosapi.go putBucketStorageClassPreProcess I", i+1, t.Errorf, tCase.code,
			code)
		if code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go putBucketStorageClassPreProcess II", i+1, t.Errorf,
				tCase.bucketName, bucketName)
		}
	}
}

type putBucketStorageClassType struct {
	bosPath      string
	storageClass string
}

func TestPutBucketStorageClass(t *testing.T) {
	testCases := []putBucketStorageClassType{
		// 1
		putBucketStorageClassType{
			bosPath:      "bos:/success",
			storageClass: "COLD",
		},
	}

	for _, tCase := range testCases {
		bosapi.PutBucketStorageClass(tCase.bosPath, tCase.storageClass)
	}
}

type putBucketStorageClassExecuteType struct {
	bucketName   string
	storageClass string
	err          string
}

func TestPutBucketStorageClassExecute(t *testing.T) {
	testCases := []putBucketStorageClassExecuteType{
		// 1
		putBucketStorageClassExecuteType{
			bucketName:   "success",
			storageClass: "COLD",
		},
		// 2
		putBucketStorageClassExecuteType{
			bucketName:   "error",
			storageClass: "COLD",
			err:          "errorCOLD",
		},
		// 3
		putBucketStorageClassExecuteType{
			bucketName:   "error1",
			storageClass: "COLD",
			err:          "error1COLD",
		},
	}

	for i, tCase := range testCases {
		err := bosapi.putBucketStorageClassExecute(tCase.bucketName, tCase.storageClass)
		util.ExpectEqual("bosapi.go putBucketStorageClassExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if tCase.err != "" {
			util.ExpectEqual("bosapi.go putBucketStorageClassExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

type getBucketStorageClassPreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestGetBucketStorageClassPreProcess(t *testing.T) {
	testCases := []getBucketStorageClassPreProcessType{
		// 1
		getBucketStorageClassPreProcessType{
			bosPath: "/liup",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		// 2
		getBucketStorageClassPreProcessType{
			bosPath: "liup/object",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		// 3
		getBucketStorageClassPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		// 4
		getBucketStorageClassPreProcessType{
			bosPath:    "bos:/bucket",
			bucketName: "bucket",
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := bosapi.getBucketStorageClassPreProcess(tCase.bosPath)
		util.ExpectEqual("bosapi.go getBucketStorageClassPreProcess I", i+1, t.Errorf, tCase.code,
			code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bosapi.go getBucketStorageClassPreProcess II", i+1, t.Errorf,
				tCase.bucketName, ret)
		}
	}
}

type getBucketStorageClassExecuteType struct {
	bucketName string
	err        string
}

func TestGetBucketStorageClassExecute(t *testing.T) {
	testCases := []getBucketStorageClassExecuteType{
		// 1
		getBucketStorageClassExecuteType{
			bucketName: "success",
		},
		// 2
		getBucketStorageClassExecuteType{
			bucketName: "error",
			err:        "error",
		},
		// 3
		getBucketStorageClassExecuteType{
			bucketName: "errorcold",
			err:        "errorcold",
		},
	}
	for i, tCase := range testCases {
		err := bosapi.getBucketStorageClassExecute(tCase.bucketName)
		util.ExpectEqual("bosapi.go getBucketStorageClassExecute I", i+1, t.Errorf, tCase.err == "",
			err == nil)
		if err != nil {
			util.ExpectEqual("bosapi.go getBucketStorageClassExecute II", i+1, t.Errorf, tCase.err,
				err.Error())
		}
	}
}

func TestGetBucketStorageClass(t *testing.T) {
	bosapi.GetBucketStorageClass("success")
}
