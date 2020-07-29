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
	"net/http"
	"os"
	// 	"runtime"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	// 	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"utils/util"
)

var (
	testBosCli          *BosCli
	bosClientWrapperVal *bosClientWrapper
	testBosHandler      *fakeCliHandler
)

type fakeCliHandler struct {
	multiDeleteDirArgVal string
	sigleDeleteArgVal    string
	utilDownlaodArgVal   string
	utilUploadFileArgVal string
}

func (h *fakeCliHandler) multiDeleteDir(bosClient bosClientInterface, bucketName,
	objectKey string) (int, error) {
	h.multiDeleteDirArgVal = bucketName + objectKey
	if bucketName == "error-delete-dir" {
		return 0, fmt.Errorf("error-delete-dir")
	}
	if val, err := strconv.Atoi(bucketName); err == nil {
		return val, nil
	} else {
		return 0, nil
	}
}

// Delete objects, retry twice.
func (h *fakeCliHandler) multiDeleteObjectsWithRetry(bosClient bosClientInterface,
	objectList []string, bucketName string) ([]api.DeleteObjectResult, error) {
	return []api.DeleteObjectResult{}, nil
}

// delete single object
func (h *fakeCliHandler) utilDeleteObject(bosClient bosClientInterface, bucketName,
	objectKey string) error {
	h.sigleDeleteArgVal = bucketName + objectKey
	if bucketName == "error-delete-single" {
		return fmt.Errorf("error-delete-single")
	}
	return nil
}

// copy single object
func (h *fakeCliHandler) utilCopyObject(srcBosClient, bosClient bosClientInterface, srcBucketName,
	srcObjectKey, dstBucketName, dstObjectKey, storageClass string, fileSize, fileMtime,
	timeOfgetObjectInfo int64, restart bool) error {
	if dstObjectKey == "copyDstObjectError" {
		return fmt.Errorf("error")
	}
	return nil
}

// download an object to local
func (h *fakeCliHandler) utilDownloadObject(bosClient bosClientInterface, srcBucketName, srcObjectKey,
	dstFilePath, downLoadTmp string, yes bool, fileSize, mtime, timeOfgetObjectInfo int64, restart bool) error {
	h.utilDownlaodArgVal = srcBucketName + srcObjectKey + dstFilePath
	if yes {
		h.utilDownlaodArgVal += "yes"
	}
	if dstFilePath == "error" {
		return fmt.Errorf("error")
	}
	if srcObjectKey == "error" {
		return fmt.Errorf("error")
	}
	if strings.HasPrefix(dstFilePath, "downerror") {
		return fmt.Errorf("error")
	}
	return nil
}

// upload a file

func (h *fakeCliHandler) utilUploadFile(bosClient bosClientInterface, srcPath, relSrcPath,
	dstBucketName, dstObjectKey, storageClass string, fileSize, fileMtime,
	timeOfgetObjectInfo int64, restart bool) error {
	h.utilUploadFileArgVal = srcPath + relSrcPath + dstBucketName + dstObjectKey + storageClass +
		strconv.FormatInt(fileSize, 10)
	if restart {
		h.utilUploadFileArgVal += "yes"
	}
	if dstBucketName == "error" {
		return fmt.Errorf("error")
	}
	return nil
}

// delete local file
func (h *fakeCliHandler) utilDeleteLocalFile(localPath string) error {
	if strings.HasSuffix(localPath, "error") {
		return fmt.Errorf("utilDeleteLocalFile error")
	}
	return nil
}

func (h *fakeCliHandler) CopySuperFile(srcBosClient, bosClient bosClientInterface, srcBucketName,
	srcObjectKey, dstBucketName, dstObjectKey, storageClass string, fileSize, mtime,
	timeOfgetObjectInfo int64, restart bool, testPrefix string) error {
	return nil
}

// check whether there is a bucket with specific name
func (h *fakeCliHandler) doesBucketExist(bosClient bosClientInterface, bucketName string) (bool,
	error) {
	if bucketName == "notExist" {
		return false, nil
	} else if bucketName == "error" {
		return false, fmt.Errorf("error")
	}
	return true, nil
}

type fakeBosClientForBos struct {
	opType              string
	results             []*api.ListObjectsResult
	objectMeta          *api.GetObjectMetaResult
	DeleteBucketName    string
	makeBucketName      string
	GetObjectMetaArgVal string
}

func (b *fakeBosClientForBos) HeadBucket(bucket string) error {
	if bucket == "exist" {
		return nil
	}
	if bucket == "notexist" {
		return &bce.BceServiceError{
			StatusCode: http.StatusNotFound,
		}
	}
	if bucket == "error" {
		return fmt.Errorf("error")
	}

	if bucket == "forbidden" {
		return &bce.BceServiceError{
			StatusCode: http.StatusForbidden,
		}
	}
	return fmt.Errorf("unknow error")
}

// Fake ListBuckets - list all buckets
func (b *fakeBosClientForBos) ListBuckets() (*api.ListBucketsResult, error) {
	if b.opType == "error" {
		return nil, fmt.Errorf("test")
	}
	locations := []string{"bj", "gz", "su"}
	ret := &api.ListBucketsResult{}
	if num, err := strconv.Atoi(b.opType); err == nil {
		for i := 1; i <= num; i++ {
			ret.Buckets = append(ret.Buckets, api.BucketSummaryType{
				Name:         "bucket" + strconv.Itoa(i),
				Location:     locations[i%3],
				CreationDate: util.TranTimestamptoLocalTime(int64(i*1000+23457), BOS_TIME_FORMT),
			})
		}
		return ret, nil
	} else {
		return nil, fmt.Errorf("error: %s", b.opType)
	}
}

// Fake PutBucket - create a new bucket
func (b *fakeBosClientForBos) PutBucket(bucket string) (string, error) {
	b.makeBucketName = bucket
	if bucket == "error" {
		return "", fmt.Errorf("test")
	}
	return "bj", nil
}

// Fake GetBucketLocation - get the location fo the given bucket
func (b *fakeBosClientForBos) GetBucketLocation(bucket string) (string, error) {
	return "", fmt.Errorf("test")
}

// Fake ListObjects - list all objects of the given bucket
func (b *fakeBosClientForBos) ListObjects(bucket string, args *api.ListObjectsArgs) (
	*api.ListObjectsResult, error) {
	var (
		marker int
		err    error
	)
	if bucket == "error" {
		ret := bucket + args.Prefix
		if args.Delimiter == "" {
			ret += "recursive"
		}
		return nil, fmt.Errorf(ret)
	}
	if args.Marker == "" {
		marker, err = strconv.Atoi(bucket)
	} else {
		marker, err = strconv.Atoi(args.Marker)
	}
	if err != nil {
		return nil, err
	}
	if marker < len(b.results) {
		if args.Delimiter == "" {
			return &api.ListObjectsResult{
				Contents:    b.results[marker].Contents,
				IsTruncated: b.results[marker].IsTruncated,
				NextMarker:  b.results[marker].NextMarker,
			}, nil
		} else {
			return b.results[marker], nil
		}
	}
	return nil, fmt.Errorf("Error in list objects")
}

// Fake DeleteBucket - delete a empty bucket
func (b *fakeBosClientForBos) DeleteBucket(bucket string) error {
	b.DeleteBucketName = bucket
	if bucket == "error-delete-bucket" {
		return fmt.Errorf("error-delete-bucket")
	}
	return nil

}

// Fake BasicGeneratePresignedUrl  generate an authorization url with expire time
func (b *fakeBosClientForBos) BasicGeneratePresignedUrl(bucket string, object string,
	expireInSeconds int) string {
	return ""
}

// Fake DeleteMultipleObjectsFromKeyList - delete a list of objects with given key string array
func (b *fakeBosClientForBos) DeleteMultipleObjectsFromKeyList(bucket string,
	keyList []string) (*api.DeleteMultipleObjectsResult, error) {
	retryNum := 1
	if strings.HasSuffix(keyList[0], "retry") {
		retryNum = 2
	}
	ret := &api.DeleteMultipleObjectsResult{}
	haveError := false
	for _, val := range keyList {
		if strings.HasPrefix(val, "TFAILED") {
			components := strings.Split(val, "-")
			// retry success delete object
			if components[1] == "2" && retryNum >= 2 {
				continue
			}
			val += "retry"
			ret.Errors = append(ret.Errors, api.DeleteObjectResult{
				Key: val, Code: components[2], Message: components[3]})
			haveError = true
		}
	}
	if haveError {
		return ret, nil
	}
	return nil, nil
}

// Fake DeleteObject - delete the given object
func (b *fakeBosClientForBos) DeleteObject(bucket, object string) error {
	if bucket == "error" {
		return fmt.Errorf(bucket + object)
	}
	return nil
}

// Fake GetObjectMeta
func (b *fakeBosClientForBos) GetObjectMeta(bucket, object string) (*api.GetObjectMetaResult,
	error) {

	b.GetObjectMetaArgVal = bucket + object
	if object == "404" {
		return nil, &bce.BceServiceError{
			StatusCode: 404,
		}
	}
	if b.objectMeta == nil {
		return nil, fmt.Errorf("Error in list objects")
	}
	return b.objectMeta, nil
}

// Fake Copy Object
func (b *fakeBosClientForBos) CopyObject(bucket, object, srcBucket, srcObject string,
	args *api.CopyObjectArgs) (*api.CopyObjectResult, error) {
	if srcBucket == "error" {
		return nil, fmt.Errorf("error")
	}
	if bucket == "" || object == "" || srcBucket == "" || srcObject == "" {
		return nil, fmt.Errorf("args error")
	}
	return nil, nil
}

// Fake of BasiGetObjectToFile
func (b *fakeBosClientForBos) BasicGetObjectToFile(bucket, object, localPath string) error {
	if bucket == "success" && object == "a/b/c" {
		return nil
	}
	return fmt.Errorf(bucket + object + localPath)
}

// Fake of PutObjectFromFile
func (b *fakeBosClientForBos) PutObjectFromFile(bucket, object, fileName string,
	args *api.PutObjectArgs) (string, error) {
	if fileName == "success" {
		return "", nil
	}
	return "", fmt.Errorf("smail" + fileName + bucket + object + args.StorageClass)
}

// Fake of UploadSuperFile
func (b *fakeBosClientForBos) UploadSuperFile(bucket, object, fileName, storageClass string) error {
	if fileName == "success" {
		return nil
	}
	return fmt.Errorf("big" + fileName + bucket + object + storageClass)
}

// Fake of BasicUploadPart
func (b *fakeBosClientForBos) BasicUploadPart(bucket, object, uploadId string, partNumber int,
	content *bce.Body) (string, error) {
	return "", fmt.Errorf("Not support")
}

func (b *fakeBosClientForBos) UploadPartFromBytes(bucket, object, uploadId string, partNumber int,
	content []byte, args *api.UploadPartArgs) (string, error) {
	return "", fmt.Errorf("Not support")
}

// Fake of GetBucketStorageclass
func (b *fakeBosClientForBos) UploadPartCopy(bucket, object, srcBucket, srcObject, uploadId string,
	partNumber int, args *api.UploadPartCopyArgs) (*api.CopyObjectResult, error) {
	return nil, fmt.Errorf("Not support")
}

// Fake of GetBucketStorageclass
func (b *fakeBosClientForBos) InitiateMultipartUpload(bucket, object, contentType string,
	args *api.InitiateMultipartUploadArgs) (*api.InitiateMultipartUploadResult, error) {
	return nil, fmt.Errorf("Not support")
}

func (b *fakeBosClientForBos) AbortMultipartUpload(bucket, object, uploadId string) error {
	return fmt.Errorf("Not support")
}

func (b *fakeBosClientForBos) CompleteMultipartUploadFromStruct(bucket, object, uploadId string,
	parts *api.CompleteMultipartUploadArgs,
	meta map[string]string) (*api.CompleteMultipartUploadResult, error) {

	return nil, fmt.Errorf("Not support")
}

func (b *fakeBosClientForBos) GetObject(bucket, object string, responseHeaders map[string]string,
	ranges ...int64) (*api.GetObjectResult, error) {

	return nil, fmt.Errorf("Not support")
}

func init() {
	bosClientForBos := &fakeBosClientForBos{
		results: []*api.ListObjectsResult{
			&api.ListObjectsResult{
				CommonPrefixes: []api.PrefixType{
					api.PrefixType{
						Prefix: "a/dir/",
					},
					api.PrefixType{
						Prefix: "a/dir2/",
					},
				},
				Contents: []api.ObjectSummaryType{
					api.ObjectSummaryType{
						Key:          "a/b",
						LastModified: "2006-01-02T15:04:05Z",
						Size:         100,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "a/c",
						LastModified: "2016-11-02T15:04:05Z",
						Size:         200,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "a/d",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
				},
				IsTruncated: true,
				NextMarker:  "1",
			},
			&api.ListObjectsResult{
				CommonPrefixes: []api.PrefixType{
					api.PrefixType{
						Prefix: "a/eir/",
					},
					api.PrefixType{
						Prefix: "a/eir2/",
					},
				},
				Contents: []api.ObjectSummaryType{
					api.ObjectSummaryType{
						Key:          "a/f",
						LastModified: "2006-01-02T15:04:05Z",
						Size:         100,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "a/g",
						LastModified: "2016-11-02T15:04:05Z",
						Size:         200,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "a/h",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
				},
				IsTruncated: false,
			},
		},
	}

	if err := initConfig(); err != nil {
		os.Exit(1)
	}

	testBosCli = NewBosCli()
	testBosCli.bosClient = bosClientForBos
	testBosHandler = &fakeCliHandler{}
	testBosCli.handler = testBosHandler
	bosClientTemp, err := buildBosClient("", "", "", bceconf.CredentialProvider,
		bceconf.ServerConfigProvider)
	if err != nil {
		fmt.Printf("create new bos.Client failed")
		os.Exit(1)
	}
	bosClientWrapperVal = &bosClientWrapper{bosClient: bosClientTemp}
}

func TestNewBosCli(t *testing.T) {
	ret := NewBosCli()
	util.ExpectEqual("bos.go TestNewBosCli I", 1, t.Errorf, false, ret == nil)
}

type genSignedUrlPreProcessType struct {
	bosPath        string
	expires        int
	haveSetExpires bool
	out            *genSignedUrlArgs
	code           BosCliErrorCode
}

func TestGenSignedUrlPreProcess(t *testing.T) {
	testCases := []genSignedUrlPreProcessType{
		genSignedUrlPreProcessType{
			bosPath:        "liupeng-bj/xxx",
			expires:        10,
			haveSetExpires: true,
			out: &genSignedUrlArgs{
				bucketName: "liupeng-bj",
				objectKey:  "xxx",
				expires:    10,
			},
			code: BOSCLI_OK,
		},
		genSignedUrlPreProcessType{
			bosPath: "bos:/liupeng-bj/xxx",
			out: &genSignedUrlArgs{
				bucketName: "liupeng-bj",
				objectKey:  "xxx",
				expires:    1800,
			},
			code: BOSCLI_OK,
		},
		genSignedUrlPreProcessType{
			bosPath:        "/liupeng-bj/xxx",
			expires:        -2,
			haveSetExpires: true,
			code:           BOSCLI_BOSPATH_IS_INVALID,
		},
		genSignedUrlPreProcessType{
			bosPath:        "bos://liupeng-bj/xxx",
			expires:        2,
			haveSetExpires: true,
			out: &genSignedUrlArgs{
				bucketName: "liupeng-bj",
				objectKey:  "xxx",
				expires:    2,
			},
			code: BOSCLI_OK,
		},
		genSignedUrlPreProcessType{
			bosPath:        "bos:/",
			expires:        -2,
			haveSetExpires: true,
			code:           BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		genSignedUrlPreProcessType{
			bosPath:        "bos:/liupeng-bj/",
			expires:        -2,
			haveSetExpires: true,
			code:           BOSCLI_OBJECTKEY_IS_EMPTY,
		},
		genSignedUrlPreProcessType{
			bosPath:        "",
			expires:        -2,
			haveSetExpires: true,
			code:           BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		genSignedUrlPreProcessType{
			bosPath:        "bos://liupeng-bj/xxx",
			expires:        -2,
			haveSetExpires: true,
			code:           BOSCLI_EXPIRE_LESS_NONE,
		},
	}
	for i, tCase := range testCases {
		ret, code := testBosCli.genSignedUrlPreProcess(tCase.bosPath, tCase.expires,
			tCase.haveSetExpires)
		util.ExpectEqual("tools.go genSignedUrlPreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("tools.go genSignedUrlPreProcess II", i+1, t.Errorf, tCase.out, ret)
		}
		if tCase.code != BOSCLI_OK && code == BOSCLI_OK {
			t.Logf("bucket '%s' object '%s' expires %d", ret.bucketName, ret.objectKey, ret.expires)
		}
	}
}

func NewFakeLocalListIterator(files []listFileResult) *fakeLocalListIterator {
	fake := &fakeLocalListIterator{
		files:     files,
		filesChan: make(chan listFileResult, 100),
	}
	return fake
}

type fakeLocalListIterator struct {
	files     []listFileResult
	filesChan chan listFileResult
}

func (f *fakeLocalListIterator) walkLocal() {
	for _, val := range f.files {
		f.filesChan <- val

	}
}

func (f *fakeLocalListIterator) next() (*fileDetail, error) {
	return nil, nil
}

type genSignedUrlType struct {
	bosPath        string
	expires        int
	haveSetExpires bool
}

func TestGenSignedUrl(t *testing.T) {
	testCases := []genSignedUrlPreProcessType{
		genSignedUrlPreProcessType{
			bosPath:        "liupeng-bj/xxx",
			expires:        10,
			haveSetExpires: true,
		},
		genSignedUrlPreProcessType{
			bosPath: "bos:/liupeng-bj/xxx",
		},
	}
	for _, tCase := range testCases {
		testBosCli.GenSignedUrl(tCase.bosPath, tCase.expires, tCase.haveSetExpires)
	}
}

type listType struct {
	bosPath   string
	opType    string
	all       bool
	recursive bool
	summary   bool
}

func TestList(t *testing.T) {
	testCases := []listType{
		listType{
			bosPath: "",
			opType:  "10",
		},
		listType{
			bosPath: "",
			opType:  "0",
		},
		listType{
			bosPath: "",
			opType:  "10",
		},
		listType{
			bosPath:   "0",
			all:       true,
			recursive: true,
		},
		listType{
			bosPath:   "0",
			all:       false,
			recursive: true,
		},
		listType{
			bosPath:   "0",
			all:       false,
			recursive: false,
		},
		listType{
			bosPath:   "bos:/0",
			all:       false,
			recursive: false,
		},
	}
	for _, tCase := range testCases {
		if fakeClient, ok := testBosCli.bosClient.(*fakeBosClientForBos); ok {
			fakeClient.opType = tCase.opType
		} else {
			t.Errorf("List: bosClient is not fakeBosClientForBos")
			continue
		}
		testBosCli.List(tCase.bosPath, tCase.all, tCase.recursive, tCase.summary)
	}
}

type listBucketsType struct {
	opType  string
	sum     int
	needSum bool
	isSuc   bool
}

func TestBosListBuckets(t *testing.T) {
	testCases := []listBucketsType{
		listBucketsType{
			opType: "error",
			isSuc:  false,
		},
		listBucketsType{
			opType:  "10",
			sum:     10,
			needSum: true,
			isSuc:   true,
		},
		listBucketsType{
			opType: "0",
			sum:    0,
			isSuc:  true,
		},
	}
	for i, tCase := range testCases {
		if fakeClient, ok := testBosCli.bosClient.(*fakeBosClientForBos); ok {
			fakeClient.opType = tCase.opType
		} else {
			t.Errorf("bosClient is not fakeBosClientForBos")
			continue
		}
		ret, err := testBosCli.listBuckets(tCase.needSum)
		util.ExpectEqual("tools.go listBuckets I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("tools.go listBuckets II", i+1, t.Errorf, tCase.sum, ret)
		}
	}
}

type bosListObjectsType struct {
	bucketName string
	objectKey  string
	all        bool
	recursive  bool
	summary    bool
	out        string
	isSuc      bool
}

func TestBosListObjects(t *testing.T) {
	testCases := []bosListObjectsType{
		bosListObjectsType{
			bucketName: "0",
			objectKey:  "",
			all:        true,
			recursive:  true,
			summary:    true,
			isSuc:      true,
		},
		bosListObjectsType{
			bucketName: "0",
			objectKey:  "",
			all:        false,
			recursive:  false,
			summary:    true,
			isSuc:      true,
		},
		bosListObjectsType{
			bucketName: "0",
			objectKey:  "",
			all:        false,
			recursive:  true,
			summary:    true,
			isSuc:      true,
		},
		bosListObjectsType{
			bucketName: "0",
			objectKey:  "",
			all:        true,
			recursive:  false,
			isSuc:      true,
		},
		bosListObjectsType{
			bucketName: "error",
			objectKey:  "testKey",
			all:        false,
			recursive:  true,
			out:        "errortestKeyrecursive",
			isSuc:      false,
		},
	}
	for i, tCase := range testCases {
		err := testBosCli.listObjects(tCase.bucketName, tCase.objectKey, tCase.all, tCase.recursive,
			tCase.summary)
		util.ExpectEqual("tools.go listObjects I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if !tCase.isSuc {
			util.ExpectEqual("tools.go listObjects I", i+1, t.Errorf, tCase.out, err.Error())
		}
	}
}

type makeBucketPreProcessType struct {
	bucketName   string
	region       string
	yes          string
	useAuto      bool
	changeClient bool
	out          string
	code         BosCliErrorCode
}

func TestMakeBucketPreProcess(t *testing.T) {
	testCases := []makeBucketPreProcessType{
		//1
		makeBucketPreProcessType{
			bucketName: "/bucket",
			code:       BOSCLI_BOSPATH_IS_INVALID,
		},
		makeBucketPreProcessType{
			bucketName: "",
			code:       BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		//3
		makeBucketPreProcessType{
			bucketName: "bos:/bucket/key",
			code:       BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
		makeBucketPreProcessType{
			bucketName: "bos:/bucket",
			out:        "bucket",
			region:     "bj",
			code:       BOSCLI_OK,
		},
		//5
		makeBucketPreProcessType{
			bucketName: "bucket",
			out:        "bucket",
			code:       BOSCLI_OK,
		},
		makeBucketPreProcessType{
			bucketName: "liupen-gz",
			out:        "liupen-gz",
			region:     "gz",
			useAuto:    false,
			yes:        "yes",
			code:       BOSCLI_OK,
		},
		//7
		makeBucketPreProcessType{
			bucketName: "liupen-gz",
			region:     "gz",
			useAuto:    false,
			yes:        "no",
			code:       BOSCLI_OPRATION_CANCEL,
		},
		makeBucketPreProcessType{
			bucketName: "liupen-gz",
			region:     "gz",
			useAuto:    true,
			code:       BOSCLI_INTERNAL_ERROR,
		},
		//9
		makeBucketPreProcessType{
			bucketName:   "liupen-gz/",
			out:          "liupen-gz",
			region:       "gz",
			useAuto:      true,
			changeClient: true,
			code:         BOSCLI_OK,
		},
		makeBucketPreProcessType{
			bucketName: "bos://bucket",
			out:        "bucket",
			region:     "bj",
			code:       BOSCLI_OK,
		},
	}
	temp := bceconf.ServerConfigProvider
	// generate server config provider
	serverConfigFileProvider, err := bceconf.NewFileServerConfigProvider("./config/config")
	if err != nil {
		t.Errorf("makeBucket init NewFileServerConfigProvider failed")
		return
	}
	defaServerConfigProvider, err := bceconf.NewDefaultServerConfigProvider()
	if err != nil {
		t.Errorf("makeBucket init NewDefaultServerConfigProvider failed")
		return
	}
	bceconf.ServerConfigProvider = bceconf.NewChainServerConfigProvider(
		[]bceconf.ServerConfigProviderInterface{serverConfigFileProvider,
			defaServerConfigProvider})

	stdinTemp := os.Stdin
	for i, tCase := range testCases {
		var (
			tempFileName string
			fd           *os.File
			err          error
		)
		serverConfigFileProvider.SetRegion("bj")
		if tCase.useAuto {
			ok := serverConfigFileProvider.SetUseAutoSwitchDomain("yes")
			if !ok {
				t.Errorf("makeBucket set SetUseAutoSwitchDomain failed %t", tCase.useAuto)
				continue
			}
		} else {
			ok := serverConfigFileProvider.SetUseAutoSwitchDomain("no")
			if !ok {
				t.Errorf("makeBucket set SetUseAutoSwitchDomain failed %t", tCase.useAuto)
				continue
			}
			if tCase.yes != "" {
				fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.yes)
				if err != nil {
					t.Errorf("create stdin input file filed")
					continue
				}
				defer os.Remove(tempFileName)
				os.Stdin = fd
			}
		}
		tempClient := testBosCli.bosClient
		if tCase.changeClient {
			testBosCli.bosClient = bosClientWrapperVal
		}
		ret, code := testBosCli.makeBucketPreProcess(tCase.bucketName, tCase.region)
		util.ExpectEqual("bos.go makeBucketPreProcess I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bos.go makeBucketPreProcess II", i+1, t.Errorf, tCase.out, ret)
		}
		testBosCli.bosClient = tempClient
	}
	os.Stdin = stdinTemp
	bceconf.ServerConfigProvider = temp
}

type makBucketType struct {
	bosPath    string
	region     string
	quiet      bool
	bucketName string
	isSuc      bool
}

func TestMakeBucket(t *testing.T) {
	testCases := []makBucketType{
		makBucketType{
			bosPath:    "bos-bj",
			quiet:      true,
			bucketName: "bos-bj",
			isSuc:      true,
		},
		makBucketType{
			bosPath:    "bos-bj",
			bucketName: "bos-bj",
			isSuc:      true,
		},
		makBucketType{
			bosPath:    "bos:/liupeng-bj",
			bucketName: "liupeng-bj",
			isSuc:      true,
		},
		makBucketType{
			bosPath:    "bos://liupeng-gz",
			bucketName: "liupeng-gz",
			isSuc:      true,
		},
	}
	for i, tCase := range testCases {
		testBosCli.MakeBucket(tCase.bosPath, tCase.region, tCase.quiet)
		if fakeClient, ok := testBosCli.bosClient.(*fakeBosClientForBos); ok {
			util.ExpectEqual("bos.go MakeBucket I", i+1, t.Errorf, tCase.bucketName,
				fakeClient.makeBucketName)
		} else {
			t.Logf("bos.go MakeBucket tran fake bosClient failed")
		}
	}
}

type removeBucketPreProcessType struct {
	bosPath    string
	bucketName string
	code       BosCliErrorCode
}

func TestRemoveBucketPreProcess(t *testing.T) {
	testCases := []removeBucketPreProcessType{
		removeBucketPreProcessType{
			bosPath:    "bos-bj",
			bucketName: "bos-bj",
			code:       BOSCLI_OK,
		},
		removeBucketPreProcessType{
			bosPath: "",
			code:    boscmd.CODE_INVALID_BUCKET_NAME,
		},
		removeBucketPreProcessType{
			bosPath:    "bos:/liupeng-bj",
			bucketName: "liupeng-bj",
			code:       BOSCLI_OK,
		},
		removeBucketPreProcessType{
			bosPath:    "bos://liupeng-gz",
			bucketName: "liupeng-gz",
			code:       BOSCLI_OK,
		},
		removeBucketPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		removeBucketPreProcessType{
			bosPath:    "bos:///liupeng-gz",
			bucketName: "liupeng-gz",
			code:       BOSCLI_OK,
		},
		removeBucketPreProcessType{
			bosPath: "/liupeng-gz",
			code:    boscmd.CODE_INVALID_BUCKET_NAME,
		},
		removeBucketPreProcessType{
			bosPath: "bos:///liupeng-gz/test",
			code:    BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME,
		},
	}
	for i, tCase := range testCases {
		ret, code := testBosCli.removeBucketPreProcess(tCase.bosPath)
		util.ExpectEqual("bos.go removeBucketPreProcessType I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bos.go removeBucketPreProcessType I", i+1, t.Errorf,
				tCase.bucketName, ret)
		}
	}
}

type removeBucketExecuteType struct {
	force          bool
	yes            bool
	needConfirmed  bool
	isSuc          bool
	bucketName     string
	confirmContent string
	code           BosCliErrorCode
	err            string
}

func TestRemoveBucketExecute(t *testing.T) {
	testCases := []removeBucketExecuteType{
		//1
		removeBucketExecuteType{
			bucketName: "bucket1",
			force:      true,
			yes:        true,
			code:       BOSCLI_OK,
			isSuc:      true,
		},
		removeBucketExecuteType{
			bucketName:     "bucket2",
			force:          true,
			yes:            false,
			confirmContent: "yes",
			code:           BOSCLI_OK,
			isSuc:          true,
		},
		removeBucketExecuteType{
			bucketName:     "bucket3",
			force:          true,
			yes:            false,
			confirmContent: "no",
			code:           BOSCLI_OPRATION_CANCEL,
			isSuc:          false,
		},
		removeBucketExecuteType{
			bucketName:     "bucket4",
			yes:            false,
			confirmContent: "no",
			code:           BOSCLI_OPRATION_CANCEL,
			isSuc:          false,
		},
		removeBucketExecuteType{
			bucketName:     "bucket5",
			yes:            false,
			confirmContent: "yes",
			code:           BOSCLI_OK,
			isSuc:          true,
		},
		removeBucketExecuteType{
			bucketName: "error-delete-dir",
			force:      true,
			yes:        true,
			code:       BOSCLI_EMPTY_CODE,
			isSuc:      false,
		},
		removeBucketExecuteType{
			bucketName: "error-delete-bucket",
			force:      true,
			yes:        true,
			code:       BOSCLI_EMPTY_CODE,
			isSuc:      false,
		},
		removeBucketExecuteType{
			bucketName: "error-delete-bucket",
			yes:        true,
			code:       BOSCLI_EMPTY_CODE,
			isSuc:      false,
		},
	}
	stdinTemp := os.Stdin
	for i, tCase := range testCases {
		var (
			fd           *os.File
			err          error
			tempFileName string
		)
		if !tCase.yes {
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.confirmContent)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		}
		retCode, ret := testBosCli.removeBucketExecute(tCase.bucketName, tCase.yes, tCase.force)
		util.ExpectEqual("bos.go removeBucketExecute I", i+1, t.Errorf, tCase.isSuc, ret == nil)

		fakeClient, ok := testBosCli.handler.(*fakeCliHandler)
		if !ok {
			t.Errorf("handler is not fakeCliHandler")
			continue
		}
		fakeBosClient, ok := testBosCli.bosClient.(*fakeBosClientForBos)
		if !ok {
			t.Errorf("bosClient is not fakeBosClientForBos")
			continue
		}

		if tCase.force && tCase.yes {
			util.ExpectEqual("bos.go removeBucketExecute II", i+1, t.Errorf, tCase.bucketName,
				fakeClient.multiDeleteDirArgVal)
		} else if tCase.yes {
			util.ExpectEqual("bos.go removeBucketExecute III", i+1, t.Errorf, tCase.bucketName,
				fakeBosClient.DeleteBucketName)
		}
		util.ExpectEqual("bos.go removeBucketExecute IV", i+1, t.Errorf, tCase.code, retCode)
	}
	os.Stdin = stdinTemp
}

type removeBucketType struct {
	force          bool
	yes            bool
	needConfirmed  bool
	isSuc          bool
	bucketName     string
	confirmContent string
	out            string
	err            string
}

func TestRemoveBucket(t *testing.T) {
	testCases := []removeBucketType{
		//1
		removeBucketType{
			bucketName: "bucket1",
			force:      true,
			yes:        true,
			out:        "bucket1",
			isSuc:      true,
		},
		removeBucketType{
			bucketName:     "bucket2/",
			force:          true,
			yes:            false,
			confirmContent: "yes",
			out:            "bucket2",
			isSuc:          true,
		},
		removeBucketType{
			bucketName:     "bos:/bucket3",
			force:          true,
			yes:            false,
			confirmContent: "yes",
			out:            "bucket3",
			isSuc:          true,
		},
		removeBucketType{
			bucketName:     "bos://bucket4",
			yes:            false,
			confirmContent: "yes",
			out:            "bucket4",
			isSuc:          true,
		},
		removeBucketType{
			bucketName:     "bos:/bucket5/",
			yes:            false,
			confirmContent: "yes",
			out:            "bucket5",
			isSuc:          true,
		},
		removeBucketType{
			bucketName:     "bos:/bucket6/",
			yes:            true,
			confirmContent: "yes",
			out:            "bucket6",
			isSuc:          true,
		},
	}
	stdinTemp := os.Stdin
	for i, tCase := range testCases {
		var (
			fd           *os.File
			err          error
			tempFileName string
		)
		if !tCase.yes {
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.confirmContent)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		}
		testBosCli.RemoveBucket(tCase.bucketName, tCase.force, tCase.yes, false)

		fakeClient, ok := testBosCli.handler.(*fakeCliHandler)
		if !ok {
			t.Errorf("handler is not fakeCliHandler")
			continue
		}
		fakeBosClient, ok := testBosCli.bosClient.(*fakeBosClientForBos)
		if !ok {
			t.Errorf("bosClient is not fakeBosClientForBos")
			continue
		}

		if tCase.force && tCase.yes {
			util.ExpectEqual("bos.go removeBucket II", i+1, t.Errorf, tCase.out,
				fakeClient.multiDeleteDirArgVal)
		} else if tCase.yes {
			util.ExpectEqual("bos.go removeBucket III", i+1, t.Errorf, tCase.out,
				fakeBosClient.DeleteBucketName)
		}
	}
	os.Stdin = stdinTemp
}

type removeObjectPreProcessType struct {
	bosPath    string
	recursive  bool
	bucketName string
	objectKey  string
	isDir      bool
	code       BosCliErrorCode
}

func TestRemoveObjectPreProcess(t *testing.T) {
	testCases := []removeObjectPreProcessType{
		removeObjectPreProcessType{
			bosPath: "/bos-bj",
			code:    BOSCLI_BOSPATH_IS_INVALID,
		},
		removeObjectPreProcessType{
			bosPath: "",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		removeObjectPreProcessType{
			bosPath: "bos:/liupeng-bj",
			code:    BOSCLI_RM_DIR_MUST_USE_RECURSIVE,
		},
		removeObjectPreProcessType{
			bosPath: "bos:/liupeng-bj/key/",
			code:    BOSCLI_RM_DIR_MUST_USE_RECURSIVE,
		},
		removeObjectPreProcessType{
			bosPath:    "bos://liupeng-gz/key",
			bucketName: "liupeng-gz",
			objectKey:  "key",
			code:       BOSCLI_OK,
		},
		removeObjectPreProcessType{
			bosPath: "bos://",
			code:    BOSCLI_BUCKETNAME_IS_EMPTY,
		},
		removeObjectPreProcessType{
			bosPath:    "bos://liupeng-gz/key",
			recursive:  true,
			bucketName: "liupeng-gz",
			objectKey:  "key",
			isDir:      true,
			code:       BOSCLI_OK,
		},
		removeObjectPreProcessType{
			bosPath:    "bos://liupeng-gz/",
			recursive:  true,
			bucketName: "liupeng-gz",
			isDir:      true,
			code:       BOSCLI_OK,
		},
		removeObjectPreProcessType{
			bosPath:    "bos://liupeng-gz",
			recursive:  true,
			bucketName: "liupeng-gz",
			isDir:      true,
			code:       BOSCLI_OK,
		},
		removeObjectPreProcessType{
			bosPath:    "bos://liupeng-gz/key/",
			recursive:  true,
			bucketName: "liupeng-gz",
			objectKey:  "key/",
			isDir:      true,
			code:       BOSCLI_OK,
		},
	}
	for i, tCase := range testCases {
		ret, code := testBosCli.removeObjectPreProcess(tCase.bosPath, tCase.recursive)
		util.ExpectEqual("bos.go removeObjectPreProcessType I", i+1, t.Errorf, tCase.code, code)
		if tCase.code == BOSCLI_OK {
			util.ExpectEqual("bos.go removeObjectPreProcessType II", i+1, t.Errorf,
				tCase.bucketName, ret.bucketName)
			util.ExpectEqual("bos.go removeObjectPreProcessType III", i+1, t.Errorf,
				tCase.objectKey, ret.objectKey)
			util.ExpectEqual("bos.go removeObjectPreProcessType IV", i+1, t.Errorf,
				tCase.isDir, ret.isDir)
		}
	}
}

type removeObjectExecuteType struct {
	yes            bool
	confirmContent string
	bucketName     string
	objectKey      string
	isDir          bool
	deleted        int
	isSuc          bool
}

func TestRemoveObjectExecute(t *testing.T) {
	testCases := []removeObjectExecuteType{
		//1
		removeObjectExecuteType{
			yes:        true,
			bucketName: "10",
			isDir:      true,
			deleted:    10,
			isSuc:      true,
		},
		removeObjectExecuteType{
			yes:        true,
			bucketName: "20",
			objectKey:  "20/",
			isDir:      true,
			deleted:    20,
			isSuc:      true,
		},
		removeObjectExecuteType{
			yes:        true,
			bucketName: "error-delete-dir",
			isDir:      true,
			isSuc:      false,
		},

		removeObjectExecuteType{
			confirmContent: "yes",
			bucketName:     "10",
			isDir:          true,
			deleted:        10,
			isSuc:          true,
		},
		removeObjectExecuteType{
			confirmContent: "yes",
			bucketName:     "20",
			objectKey:      "20/",
			isDir:          true,
			deleted:        20,
			isSuc:          true,
		},
		removeObjectExecuteType{
			confirmContent: "yes",
			bucketName:     "error-delete-dir",
			isDir:          true,
			isSuc:          false,
		},

		removeObjectExecuteType{
			confirmContent: "no",
			bucketName:     "10",
			isDir:          true,
			deleted:        0,
			isSuc:          true,
		},

		removeObjectExecuteType{
			yes:        true,
			bucketName: "signle",
			objectKey:  "signelKey",
			deleted:    1,
			isSuc:      true,
		},
		removeObjectExecuteType{
			yes:        true,
			bucketName: "error-delete-single",
			isSuc:      false,
		},

		removeObjectExecuteType{
			confirmContent: "yes",
			bucketName:     "signle",
			objectKey:      "signelKey",
			deleted:        1,
			isSuc:          true,
		},
		removeObjectExecuteType{
			confirmContent: "yes",
			bucketName:     "error-delete-single",
			isSuc:          false,
		},

		removeObjectExecuteType{
			confirmContent: "no",
			bucketName:     "signle",
			objectKey:      "signelKey",
			deleted:        0,
			isSuc:          true,
		},
	}
	stdinTemp := os.Stdin
	for i, tCase := range testCases {
		var (
			fd           *os.File
			err          error
			tempFileName string
		)
		if !tCase.yes {
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.confirmContent)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		}
		args := &removeObjectArgs{
			objectKey:  tCase.objectKey,
			bucketName: tCase.bucketName,
			isDir:      tCase.isDir,
		}
		ret, err := testBosCli.removeObjectExecute(args, tCase.yes)
		util.ExpectEqual("bos.go removeObjectExecute I", i+1, t.Errorf, tCase.isSuc, err == nil)
		fakeClient, ok := testBosCli.handler.(*fakeCliHandler)
		if !ok {
			t.Errorf("handler is not fakeCliHandler")
			continue
		}

		if tCase.isSuc {
			util.ExpectEqual("bos.go removeObjectExecute II", i+1, t.Errorf, tCase.deleted, ret)
		}
		if tCase.confirmContent != "no" {
			if tCase.isDir {
				util.ExpectEqual("bos.go removeObjectExecute III", i+1, t.Errorf,
					tCase.bucketName+tCase.objectKey, fakeClient.multiDeleteDirArgVal)
			} else {
				util.ExpectEqual("bos.go removeObjectExecute IV", i+1, t.Errorf,
					tCase.bucketName+tCase.objectKey, fakeClient.sigleDeleteArgVal)
			}
		}
	}
	os.Stdin = stdinTemp
}

type removeObjectType struct {
	yes            bool
	recursive      bool
	isDir          bool
	bosPath        string
	out            string
	confirmContent string
}

func TestRemoveObject(t *testing.T) {

	testCases := []removeObjectType{
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "bos://10",
			out:       "10",
		},
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "20",
			out:       "20",
		},
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "30/",
			out:       "30",
		},
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "40/key",
			out:       "40key",
		},
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "bos:/50/key50/",
			out:       "50key50/",
		},
		removeObjectType{
			yes:       true,
			recursive: true,
			isDir:     true,
			bosPath:   "60/key40/",
			out:       "60key40/",
		},
		removeObjectType{
			confirmContent: "yes",
			recursive:      true,
			isDir:          true,
			bosPath:        "70/key",
			out:            "70key",
		},
		removeObjectType{
			confirmContent: "yes",
			bosPath:        "80/key",
			out:            "80key",
		},
		removeObjectType{
			confirmContent: "no",
			bosPath:        "bos/key90",
			out:            "boskey90",
		},
		removeObjectType{
			confirmContent: "yes",
			bosPath:        "bos:/pas/key100",
			out:            "paskey100",
		},
		removeObjectType{
			confirmContent: "yes",
			bosPath:        "bos:/signelKey/key",
			out:            "signelKeykey",
		},
		removeObjectType{
			confirmContent: "yes",
			bosPath:        "bos://signelKey/key",
			out:            "signelKeykey",
		},
	}
	stdinTemp := os.Stdin
	for i, tCase := range testCases {
		var (
			fd           *os.File
			err          error
			tempFileName string
		)
		if !tCase.yes {
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.confirmContent)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		}
		testBosCli.RemoveObject(tCase.bosPath, tCase.yes, tCase.recursive, false)
		fakeClient, ok := testBosCli.handler.(*fakeCliHandler)
		if !ok {
			t.Errorf("handler is not fakeCliHandler")
			continue
		}
		if tCase.confirmContent != "no" {
			if tCase.isDir {
				util.ExpectEqual("bos.go removeObject III", i+1, t.Errorf,
					tCase.out, fakeClient.multiDeleteDirArgVal)
			} else {
				util.ExpectEqual("bos.go removeObject IV", i+1, t.Errorf,
					tCase.out, fakeClient.sigleDeleteArgVal)
			}
		}
	}
	os.Stdin = stdinTemp
}

type copyRemoteRequestPreProcessType struct {
	srcPath       string
	dstPath       string
	storageClass  string
	recursive     bool
	srcBucketName string
	srcObjectKey  string
	dstBucketName string
	dstObjectKey  string
	isDir         bool
	code          BosCliErrorCode
	isSuc         bool
}

func TestCopyRemoteRequestPreProcess(t *testing.T) {
	testCases := []copyRemoteRequestPreProcessType{
		//1
		copyRemoteRequestPreProcessType{
			srcPath:      "",
			storageClass: "xxx",
			code:         BOSCLI_UNSUPPORT_STORAGE_CLASS,
		},
		//2
		copyRemoteRequestPreProcessType{
			srcPath: "",
			code:    BOSCLI_SRC_BUCKET_IS_EMPTY,
		},
		//3
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/key",
			dstPath: "",
			code:    BOSCLI_DST_BUCKET_IS_EMPTY,
		},
		//4
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/error/key",
			dstPath: "bos:/dstBucket/key",
			code:    BOSCLI_EMPTY_CODE,
		},
		//5
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/notExist/key",
			dstPath: "bos:/dstBucket/key",
			code:    BOSCLI_SRC_BUCKET_DONT_EXIST,
		},
		//6
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/key",
			dstPath: "bos:/error/key",
			code:    BOSCLI_EMPTY_CODE,
		},
		//7
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/key",
			dstPath: "bos:/notExist/key",
			code:    BOSCLI_DST_BUCKET_DONT_EXIST,
		},
		//8
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/key/",
			dstPath: "bos:/dstBucket/key",
			code:    BOSCLI_BATCH_COPY_DSTOBJECT_END,
		},
		//9
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "bos:/dstBucket/key",
			code:    BOSCLI_BATCH_COPY_DSTOBJECT_END,
		},
		//10
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "bos:/dstBucket/key/",
			code:    BOSCLI_BATCH_COPY_SRCOBJECT_END,
		},
		//11
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "bos:/dstBucket/",
			code:    BOSCLI_BATCH_COPY_SRCOBJECT_END,
		},
		//12
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket",
			dstPath: "bos:/dstBucekt/",
			code:    BOSCLI_BATCH_COPY_SRCOBJECT_END,
		},
		//13
		copyRemoteRequestPreProcessType{
			srcPath: "bos:/bucket",
			dstPath: "bos:/destBucket",
			code:    BOSCLI_BATCH_COPY_SRCOBJECT_END,
		},
		//14
		copyRemoteRequestPreProcessType{
			srcPath:       "bos:/bucket",
			dstPath:       "bos:/dstBucket",
			storageClass:  "STANDARD",
			recursive:     true,
			srcBucketName: "bucket",
			dstBucketName: "dstBucket",
			isDir:         true,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//15
		copyRemoteRequestPreProcessType{
			srcPath:       "bos:/bucket/key/test/",
			dstPath:       "bos:/dstBucket/xxx/",
			storageClass:  "STANDARD",
			recursive:     true,
			srcBucketName: "bucket",
			srcObjectKey:  "key/test/",
			dstBucketName: "dstBucket",
			dstObjectKey:  "xxx/",
			isDir:         true,
			code:          BOSCLI_OK,
			isSuc:         true,
		},

		// 16
		copyRemoteRequestPreProcessType{
			srcPath:       "bos:/bucket15/key1/key2",
			dstPath:       "bos:/dstBucket15",
			storageClass:  "STANDARD_IA",
			srcBucketName: "bucket15",
			srcObjectKey:  "key1/key2",
			dstBucketName: "dstBucket15",
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		// 17
		copyRemoteRequestPreProcessType{
			srcPath:       "bos:/0/key/",
			dstPath:       "bos:/bucket/",
			recursive:     true,
			storageClass:  "STANDARD_IA",
			srcBucketName: "0",
			srcObjectKey:  "key/",
			dstBucketName: "bucket",
			code:          BOSCLI_OK,
			isSuc:         true,
			isDir:         true,
		},
		// 18
		copyRemoteRequestPreProcessType{
			srcPath:       "bos:/0/key/xxx",
			dstPath:       "bos:/bucket/",
			recursive:     true,
			storageClass:  "STANDARD_IA",
			srcBucketName: "0",
			srcObjectKey:  "key/xxx",
			dstBucketName: "bucket",
			code:          BOSCLI_OK,
			isDir:         false,
			isSuc:         true,
		},
	}
	for i, tCase := range testCases {
		args, code, err := testBosCli.copyRemoteRequestPreProcess(tCase.srcPath, tCase.dstPath,
			tCase.storageClass, tCase.recursive)
		util.ExpectEqual("bos.go remote pre I", i+1, t.Errorf, tCase.code, code)
		util.ExpectEqual("bos.go remote pre II", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			if err != nil {
				t.Logf("code %s error %s", code, err.Error())
			}
			util.ExpectEqual("bos.go remote pre III", i+1, t.Errorf, tCase.srcBucketName,
				args.srcBucketName)
			util.ExpectEqual("bos.go remote pre IV", i+1, t.Errorf, tCase.dstBucketName,
				args.dstBucketName)
			util.ExpectEqual("bos.go remote pre V", i+1, t.Errorf, tCase.srcObjectKey,
				args.srcObjectKey)
			util.ExpectEqual("bos.go remote pre VI", i+1, t.Errorf, tCase.dstObjectKey,
				args.dstObjectKey)
			util.ExpectEqual("bos.go remote pre VII", i+1, t.Errorf, tCase.isDir,
				args.srcIsDir)
		}
	}
}

type copyObjectList struct {
	key          string
	path         string
	storageClass string
}

func getTestCopyList(bucketName, objectKey string) ([]copyObjectList, error) {
	ret := []copyObjectList{}

	client, err := initBosClientForBucket("", "", bucketName)
	if err != nil {
		return nil, err
	}
	objectLists := NewObjectListIterator(client, nil, bucketName, objectKey, "", true, true, true,
		false, 1000)

	for {
		listResult, err := objectLists.next()
		if err != nil {
			return nil, err
		}
		if listResult.ended {
			break
		}
		if listResult.isDir {
			continue
		}

		object := listResult.file
		if object.key == "" {
			continue
		}
		ret = append(ret, copyObjectList{
			key:          object.key,
			path:         object.path,
			storageClass: object.storageClass,
		})
	}
	return ret, nil
}

type copyObjectExecuteType struct {
	srcBucket    string
	srcObject    string
	dstBucket    string
	dstObject    string
	storageClass string
	isDir        bool
	restart      bool
	copied       int
	setCopied    bool
	isSuc        bool
}

func TestCopyObjectExecute(t *testing.T) {
	testCases := []copyObjectExecuteType{
		//1
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "progress",
			dstBucket: "dstBucket",
			copied:    0,
			isDir:     false,
			isSuc:     false,
		},
		//2
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "progress/",
			dstBucket: "dstBucket",
			isDir:     true,
			isSuc:     true,
		},
		//3
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "progress/",
			dstBucket: "dstBucket",
			dstObject: "key/",
			isDir:     true,
			isSuc:     true,
		},
		//4
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "xxx",
			dstBucket: "dstBucket",
			dstObject: "key/",
			isDir:     false,
			isSuc:     false,
		},
		//5
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "xxx",
			dstBucket: "dstBucket",
			dstObject: "key",
			isDir:     false,
			isSuc:     false,
		},
		//6
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "bce",
			dstBucket: "dstBucket",
			dstObject: "key",
			isDir:     false,
			copied:    1,
			setCopied: true,
			isSuc:     true,
		},
		//7
		copyObjectExecuteType{
			srcBucket:    "cli-test",
			srcObject:    "progress/",
			dstBucket:    "cli-test",
			dstObject:    "progress/",
			storageClass: "STANDARD",
			isDir:        true,
			copied:       0,
			setCopied:    true,
			isSuc:        true,
		},
		//8
		copyObjectExecuteType{
			srcBucket: "cli-test",
			srcObject: "progress/",
			dstBucket: "bucket",
			dstObject: "copyDstObjectError",
			isDir:     true,
			copied:    0,
			setCopied: true,
			isSuc:     true,
		},
		//9
		copyObjectExecuteType{
			srcBucket: "cli-testxxxxx",
			srcObject: "progress",
			dstBucket: "bucket",
			dstObject: "copyDstObjectError",
			isSuc:     false,
		},
	}
	for i, tCase := range testCases {
		args := &copyBetweenRemoteArgs{
			srcBucketName: tCase.srcBucket,
			srcObjectKey:  tCase.srcObject,
			dstBucketName: tCase.dstBucket,
			dstObjectKey:  tCase.dstObject,
			srcIsDir:      tCase.isDir,
		}
		ret, _, err := testBosCli.copyObjectExecute(args, tCase.storageClass, true)
		util.ExpectEqual("bos.go remote exe I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			oCopied := 0
			if tCase.setCopied {
				oCopied = tCase.copied
			} else {
				objects, err := getTestCopyList(tCase.srcBucket, tCase.srcObject)
				if err != nil {
					t.Errorf("get object num failed: %s %s", tCase.srcBucket, tCase.srcObject)
				}
				oCopied = len(objects)
			}
			util.ExpectEqual("bos.go remote exe II", i+1, t.Errorf, oCopied, ret.successed)
			t.Logf("id %d object Num %d copied %d", i+1, oCopied, ret.successed)
		}
		if err != nil {
			t.Logf("id %d copy error: %s", i+1, err)
		}

	}
}

type copyRemoteType struct {
	srcPath      string
	dstPath      string
	storageClass string
	recursive    bool
	isSuc        bool
}

func TestCopyBetweenRemote(t *testing.T) {
	testCases := []copyRemoteType{
		//1
		copyRemoteType{
			srcPath:   "cli-test/",
			dstPath:   "bucekt",
			recursive: true,
			isSuc:     true,
		},
		//2
		copyRemoteType{
			srcPath: "cli-test/bce",
			dstPath: "bucekt/bce",
			isSuc:   true,
		},
		//3
		copyRemoteType{
			srcPath: "bos:/cli-test/bce",
			dstPath: "bucekt/bce",
			isSuc:   true,
		},
		//4
		copyRemoteType{
			srcPath: "bos://cli-test/bce",
			dstPath: "bos:/bucekt/bce",
			isSuc:   true,
		},
	}
	for i, tCase := range testCases {
		retCode, _ := testBosCli.copyBetweenRemote(tCase.srcPath, tCase.dstPath, tCase.storageClass,
			tCase.recursive, true)
		util.ExpectEqual("bos.go copyBetweenRemote", i+1, t.Errorf, tCase.isSuc, retCode == BOSCLI_OK)
	}
}

type copyDownloadPreProcessType struct {
	srcPath       string
	dstPath       string
	recursive     bool
	srcBucketName string
	srcObjectKey  string
	isDir         bool
	code          BosCliErrorCode
	isSuc         bool
}

func TestCopyDownloadPreProcess(t *testing.T) {
	local_dir := "./test_dir"
	util.TryMkdir(local_dir)
	defer os.RemoveAll(local_dir)
	testCases := []copyDownloadPreProcessType{
		//1
		copyDownloadPreProcessType{
			srcPath: "",
			code:    BOSCLI_SRC_BUCKET_IS_EMPTY,
			isSuc:   false,
		},
		//2
		copyDownloadPreProcessType{
			srcPath: "notExist",
			code:    BOSCLI_SRC_BUCKET_DONT_EXIST,
			isSuc:   false,
		},
		//3
		copyDownloadPreProcessType{
			srcPath: "bos:/notExist",
			code:    BOSCLI_SRC_BUCKET_DONT_EXIST,
			isSuc:   false,
		},
		//4
		copyDownloadPreProcessType{
			srcPath: "bos:/notExist/",
			code:    BOSCLI_SRC_BUCKET_DONT_EXIST,
			isSuc:   false,
		},
		//5
		copyDownloadPreProcessType{
			srcPath: "bos:/error/",
			code:    BOSCLI_EMPTY_CODE,
			isSuc:   false,
		},
		//6
		copyDownloadPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "/root/",
			code:    BOSCLI_DIR_IS_NOT_WRITABLE,
			isSuc:   false,
		},
		//7
		copyDownloadPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "./root/",
			code:    BOSCLI_BATCH_DOWNLOAD_SRCOBJECT_END,
			isSuc:   false,
		},
		//7
		copyDownloadPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: "./bos_test.go",
			code:    BOSCLI_CANT_DOWNLOAD_FILES_TO_FILE,
			isSuc:   false,
		},

		//8
		copyDownloadPreProcessType{
			srcPath: "bos:/bucket/",
			dstPath: local_dir,
			code:    BOSCLI_BATCH_DOWNLOAD_SRCOBJECT_END,
			isSuc:   false,
		},
		//9
		copyDownloadPreProcessType{
			srcPath:       "bos:/bucket/",
			dstPath:       local_dir,
			recursive:     true,
			srcBucketName: "bucket",
			srcObjectKey:  "",
			isDir:         true,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//10
		copyDownloadPreProcessType{
			srcPath:       "bos:/bucket/bce",
			dstPath:       local_dir,
			srcBucketName: "bucket",
			srcObjectKey:  "bce",
			isDir:         false,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//11
		copyDownloadPreProcessType{
			srcPath:       "bos:/0/",
			dstPath:       "./",
			srcBucketName: "0",
			recursive:     true,
			srcObjectKey:  "",
			isDir:         true,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//12
		copyDownloadPreProcessType{
			srcPath:       "bos://bucket/bce",
			dstPath:       local_dir,
			srcBucketName: "bucket",
			srcObjectKey:  "bce",
			isDir:         false,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
	}
	for i, tCase := range testCases {
		args, code, err := testBosCli.copyDownloadPreProcess(tCase.srcPath, tCase.dstPath,
			tCase.recursive)

		util.ExpectEqual("bos.go down pre I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("bos.go down pre II", i+1, t.Errorf, tCase.code, code)
		if tCase.isSuc {
			if err != nil {
				t.Logf("%s %s", code, err)
			}
			util.ExpectEqual("bos.go down pre III", i+1, t.Errorf, tCase.srcBucketName,
				args.srcBucketName)
			util.ExpectEqual("bos.go down pre IV", i+1, t.Errorf, tCase.srcObjectKey,
				args.srcObjectKey)
			util.ExpectEqual("bos.go down pre IV", i+1, t.Errorf, tCase.isDir,
				args.srcIsDir)
		}
	}
}

type copyDownloadExecuteType struct {
	srcBucketName string
	srcObjectKey  string
	dstPath       string
	downLoadTmp   string
	out           string
	haveSep       bool
	isDir         bool
	downed        int
	code          BosCliErrorCode
	isSuc         bool
}

var (
	tempFakeBosClient = &fakeBosClientForBos{
		objectMeta: &api.GetObjectMetaResult{
			api.ObjectMeta{
				LastModified:  "Wed, 06 Apr 2016 06:34:40 GMT",
				ContentLength: 100,
				StorageClass:  "STANDARD",
			},
		},
		results: []*api.ListObjectsResult{
			&api.ListObjectsResult{
				Contents: []api.ObjectSummaryType{
					api.ObjectSummaryType{
						Key:          "key/a/b",
						LastModified: "2006-01-02T15:04:05Z",
						Size:         100,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "key/a/c",
						LastModified: "2016-11-02T15:04:05Z",
						Size:         200,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "key/a/d",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
					api.ObjectSummaryType{
						Key:          "key/a/f",
						LastModified: "2006-01-02T15:04:05Z",
						Size:         100,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "key/a/g",
						LastModified: "2016-11-02T15:04:05Z",
						Size:         200,
						StorageClass: "",
					},
					api.ObjectSummaryType{
						Key:          "key/a/h",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
					api.ObjectSummaryType{
						Key:          "key/a/error",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
					api.ObjectSummaryType{
						Key:          "key/a/i",
						LastModified: "2017-11-02T15:04:05Z",
						Size:         300,
						StorageClass: "STANDARD",
					},
				},
				IsTruncated: false,
			},
		},
	}
)

func TestCopyDownloadExecute(t *testing.T) {

	tempClientBos := testBosCli.bosClient
	testBosCli.bosClient = tempFakeBosClient
	testCases := []copyDownloadExecuteType{
		copyDownloadExecuteType{
			srcBucketName: "error",
			srcObjectKey:  "",
			dstPath:       "./",
			haveSep:       true,
			isDir:         true,
			isSuc:         false,
		},
		copyDownloadExecuteType{
			srcBucketName: "0",
			srcObjectKey:  "key/",
			dstPath:       "./",
			out:           "0key/a/i./a/iyes",
			haveSep:       true,
			isDir:         true,
			downed:        8,
			isSuc:         true,
		},
		copyDownloadExecuteType{
			srcBucketName: "0",
			srcObjectKey:  "key/",
			dstPath:       "./bcetest",
			out:           "0key/a/i./bcetest/a/iyes",
			haveSep:       true,
			isDir:         true,
			downed:        8,
			isSuc:         true,
		},
		copyDownloadExecuteType{
			srcBucketName: "0",
			srcObjectKey:  "key/a/",
			dstPath:       "./bcetest",
			out:           "0key/a/i./bcetest/iyes",
			haveSep:       true,
			isDir:         true,
			downed:        8,
			isSuc:         true,
		},
		copyDownloadExecuteType{
			srcBucketName: "0",
			srcObjectKey:  "key/",
			dstPath:       "downerror",
			out:           "0key/a/idownerror/a/iyes",
			haveSep:       true,
			isDir:         true,
			downed:        0,
			isSuc:         true,
		},

		// single object success
		copyDownloadExecuteType{
			srcBucketName: "0",
			srcObjectKey:  "key",
			dstPath:       "./",
			out:           "0key./yes",
			downed:        1,
			isSuc:         true,
		},
		// single object success
		copyDownloadExecuteType{
			srcBucketName: "bucket",
			srcObjectKey:  "test/key",
			dstPath:       "./",
			out:           "buckettest/key./yes",
			downed:        1,
			isSuc:         true,
		},
		// single object error
		copyDownloadExecuteType{
			srcBucketName: "error",
			srcObjectKey:  "error",
			dstPath:       "./",
			isSuc:         false,
		},
	}
	for i, tCase := range testCases {
		args := &copyDownloadArgs{
			srcBucketName: tCase.srcBucketName,
			srcObjectKey:  tCase.srcObjectKey,
			srcIsDir:      tCase.isDir,
		}
		ret, _, err := testBosCli.copyDownloadExecute(args, tCase.dstPath, tCase.downLoadTmp, true, false)
		util.ExpectEqual("bos.go down exe I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("bos.go down exe II", i+1, t.Errorf, tCase.downed, ret.successed)
			util.ExpectEqual("bos.go down exe III", i+1, t.Errorf, tCase.out,
				testBosHandler.utilDownlaodArgVal)
		}
		t.Logf("want downed %d get %d", tCase.downed, ret.successed)
	}
	testBosCli.bosClient = tempClientBos
}

type copyDownloadType struct {
	srcPath     string
	dstPath     string
	downLoadTmp string
	out         string
	recursive   bool
	isSuc       bool
}

func TestCopyDownload(t *testing.T) {
	tempClientBos := testBosCli.bosClient
	testBosCli.bosClient = tempFakeBosClient
	testCases := []copyDownloadType{
		copyDownloadType{
			srcPath:   "bos:/0/key/",
			dstPath:   "./",
			out:       "0key/a/i./a/iyes",
			recursive: true,
			isSuc:     true,
		},
		copyDownloadType{
			srcPath:   "bos:/bucket/test/key",
			dstPath:   "./",
			out:       "buckettest/key./yes",
			recursive: true,
			isSuc:     true,
		},
		copyDownloadType{
			srcPath:   "bos://bucket/test/key",
			dstPath:   "./",
			out:       "buckettest/key./yes",
			recursive: true,
			isSuc:     true,
		},
	}
	for i, tCase := range testCases {
		retCode, _ := testBosCli.copyDownload(tCase.srcPath, tCase.dstPath, tCase.downLoadTmp, tCase.recursive,
			true, false)
		util.ExpectEqual("bos.go down I", i+1, t.Errorf, tCase.isSuc,
			retCode == BOSCLI_OK)
		util.ExpectEqual("bos.go down II", i+1, t.Errorf, tCase.out,
			testBosHandler.utilDownlaodArgVal)
	}
	testBosCli.bosClient = tempClientBos
}

type copyUploadPreProcessType struct {
	srcPath       string
	dstPath       string
	storageClass  string
	recursive     bool
	dstBucketName string
	dstObjectKey  string
	isDir         bool
	code          BosCliErrorCode
	isSuc         bool
}

func TestCopyUploadPreProcess(t *testing.T) {
	pathPrefix := "./test_upload_pre"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("TestCopyUploadPreProcess create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go up pre %s", err.Error())
		}
	}()

	testCases := []copyUploadPreProcessType{
		//1
		copyUploadPreProcessType{
			dstPath: "",
			code:    BOSCLI_DST_BUCKET_IS_EMPTY,
			isSuc:   false,
		},
		//2
		copyUploadPreProcessType{
			dstPath:      "bucket",
			storageClass: "xxx",
			code:         BOSCLI_UNSUPPORT_STORAGE_CLASS,
			isSuc:        false,
		},
		//3
		copyUploadPreProcessType{
			srcPath: "notExist",
			dstPath: "bucket",
			code:    boscmd.LOCAL_PATH_NOT_EXIST,
			isSuc:   false,
		},
		//4
		copyUploadPreProcessType{
			srcPath: pathPrefix,
			dstPath: "bos:/notExist",
			code:    BOSCLI_DST_BUCKET_DONT_EXIST,
			isSuc:   false,
		},
		//5
		copyUploadPreProcessType{
			srcPath: pathPrefix,
			dstPath: "bos:/error",
			code:    BOSCLI_EMPTY_CODE,
			isSuc:   false,
		},
		//6
		copyUploadPreProcessType{
			srcPath:       pathPrefix + "/",
			dstPath:       "bos:/bucket/key",
			recursive:     true,
			dstBucketName: "bucket",
			dstObjectKey:  "key/",
			isDir:         true,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//7
		copyUploadPreProcessType{
			srcPath:       pathPrefix + "/ab",
			dstPath:       "bos:/bucket/",
			recursive:     true,
			dstBucketName: "bucket",
			isDir:         false,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//8
		copyUploadPreProcessType{
			srcPath: pathPrefix,
			dstPath: "bos:/bucket",
			code:    BOSCLI_UPLOAD_SRC_CANNT_BE_DIR,
			isSuc:   false,
		},
		//9
		copyUploadPreProcessType{
			srcPath:       pathPrefix + "/ab",
			dstPath:       "bos:/bucket/",
			dstBucketName: "bucket",
			isDir:         false,
			code:          BOSCLI_OK,
			isSuc:         true,
		},
		//10
		copyUploadPreProcessType{
			srcPath: "-",
			dstPath: "bos:/bucket",
			code:    BOSCLI_DST_OBJECT_KEY_IS_EMPTY,
			isSuc:   false,
		},
		//11
		copyUploadPreProcessType{
			srcPath: "-",
			dstPath: "bos:/bucekt/xx/",
			code:    BOSCLI_UPLOAD_STREAM_TO_DIR,
			isSuc:   false,
		},
		//12
		copyUploadPreProcessType{
			srcPath: "-",
			dstPath: "bos:/bucekt/xx",
			code:    BOSCLI_UNSUPPORT_METHOD,
			isSuc:   false,
		},
	}
	for i, tCase := range testCases {
		args, code, err := testBosCli.copyUploadRequestPreProcess(tCase.srcPath, tCase.dstPath,
			tCase.storageClass, tCase.recursive)

		util.ExpectEqual("bos.go up pre I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("bos.go up pre II", i+1, t.Errorf, tCase.code, code)
		if tCase.isSuc {
			util.ExpectEqual("bos.go up pre III", i+1, t.Errorf, tCase.dstBucketName,
				args.dstBucketName)
			util.ExpectEqual("bos.go up pre IV", i+1, t.Errorf, tCase.dstObjectKey,
				args.dstObjectKey)
			util.ExpectEqual("bos.go up pre V", i+1, t.Errorf, tCase.isDir,
				args.srcIsDir)
			util.ExpectEqual("bos.go up pre VI", i+1, t.Errorf, tCase.srcPath,
				args.srcPath)
		}
	}
}

type copyUploadExecuteType struct {
	srcPath          string
	dstBucketName    string
	dstObjectKey     string
	storageClass     string
	finlObjectKey    string
	isDir            bool
	uploadFromStream bool
	uploaded         int
	code             BosCliErrorCode
	isSuc            bool
}

func TestUploadFileExecute(t *testing.T) {
	pathPrefix := "./test_upload_exe"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go up exe %s", err.Error())
		}
	}()

	testCases := []copyUploadExecuteType{
		//1
		copyUploadExecuteType{
			srcPath:       "./xxx",
			dstBucketName: "bucket",
			dstObjectKey:  "",
			isDir:         true,
			code:          BOSCLI_EMPTY_CODE,
			uploaded:      0,
			isSuc:         true,
		},
		//2
		copyUploadExecuteType{
			srcPath:       "-",
			dstBucketName: "bucket",
			dstObjectKey:  "",
			isDir:         true,
			code:          BOSCLI_UNSUPPORT_METHOD,
			isSuc:         false,
		},
		//3
		copyUploadExecuteType{
			srcPath:       pathPrefix,
			dstBucketName: "error",
			dstObjectKey:  "",
			isDir:         true,
			code:          BOSCLI_EMPTY_CODE,
			uploaded:      0,
			isSuc:         true,
		},
		//4
		copyUploadExecuteType{
			srcPath:       pathPrefix,
			dstBucketName: "bucket",
			dstObjectKey:  "",
			storageClass:  "STANDARD",
			isDir:         true,
			code:          BOSCLI_EMPTY_CODE,
			uploaded:      16,
			isSuc:         true,
		},
		//5
		copyUploadExecuteType{
			srcPath:       pathPrefix + "/aDir/",
			dstBucketName: "bucket",
			dstObjectKey:  "test/",
			isDir:         true,
			code:          BOSCLI_EMPTY_CODE,
			uploaded:      9,
			isSuc:         true,
		},
		//5
		copyUploadExecuteType{
			srcPath:       pathPrefix + "/aDir/234",
			dstBucketName: "bucket",
			dstObjectKey:  "test/",
			finlObjectKey: "test/234",
			storageClass:  "STANDARD_IA",
			isDir:         false,
			code:          BOSCLI_EMPTY_CODE,
			uploaded:      1,
			isSuc:         true,
		},
	}

	for i, tCase := range testCases {
		args := &copyUploadArges{
			srcPath:          tCase.srcPath,
			dstBucketName:    tCase.dstBucketName,
			dstObjectKey:     tCase.dstObjectKey,
			srcIsDir:         tCase.isDir,
			uploadFromStream: tCase.uploadFromStream,
		}
		ret, code, err := testBosCli.uploadFileExecute(args, tCase.srcPath, tCase.storageClass,
			true)

		util.ExpectEqual("bos.go up exe I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("bos.go up exe I", i+1, t.Errorf, tCase.code, code)
		if tCase.isSuc {
			util.ExpectEqual("bos.go up exe II", i+1, t.Errorf, tCase.uploaded, ret.successed)
			if !tCase.isDir {
				absSrcPath, _ := util.Abs(tCase.srcPath)
				relSrcPath, _ := filepath.EvalSymlinks(absSrcPath)
				fileSize, _ := util.GetSizeOfFile(relSrcPath)
				out := absSrcPath + relSrcPath + tCase.dstBucketName + tCase.finlObjectKey +
					tCase.storageClass + strconv.FormatInt(fileSize, 10)
				out += "yes"
				util.ExpectEqual("bos.go upexe III", i+1, t.Errorf, out,
					testBosHandler.utilUploadFileArgVal)
			}
		}
	}
}

type copyUploadType struct {
	srcPath        string
	dstPath        string
	dstBucketName  string
	finalObjectKey string
	storageClass   string
	recursive      bool
	isSuc          bool
}

func TestCopyUpload(t *testing.T) {
	pathPrefix := "test_copy_upload"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go up %s", err.Error())
		}
	}()

	tempClientBos := testBosCli.bosClient
	testBosCli.bosClient = tempFakeBosClient
	testCases := []copyUploadType{
		copyUploadType{
			srcPath:   pathPrefix,
			dstPath:   "liupeng-bj",
			recursive: true,
			isSuc:     false,
		},
		copyUploadType{
			srcPath:        pathPrefix + "/aDir/234",
			dstPath:        "bos:/bucket/key",
			dstBucketName:  "bucket",
			finalObjectKey: "key",
			isSuc:          true,
		},
		copyUploadType{
			srcPath:        pathPrefix + "/aDir/234",
			dstPath:        "bos:/bucket/key/",
			dstBucketName:  "bucket",
			finalObjectKey: "key/234",
			isSuc:          true,
		},
		copyUploadType{
			srcPath:        pathPrefix + "/aDir/234",
			dstPath:        "bos://bucket/key/",
			dstBucketName:  "bucket",
			finalObjectKey: "key/234",
			isSuc:          true,
		},
	}
	for i, tCase := range testCases {
		retCode, _ := testBosCli.copyUpload(tCase.srcPath, tCase.dstPath, tCase.storageClass, tCase.recursive,
			true)

		util.ExpectEqual("bos.go copyUpload I", i+1, t.Errorf, tCase.isSuc,
			retCode == BOSCLI_OK)

		if !tCase.recursive {
			absSrcPath, _ := util.Abs(tCase.srcPath)
			relSrcPath, _ := filepath.EvalSymlinks(absSrcPath)
			fileSize, _ := util.GetSizeOfFile(relSrcPath)
			out := absSrcPath + relSrcPath + tCase.dstBucketName + tCase.finalObjectKey +
				tCase.storageClass + strconv.FormatInt(fileSize, 10)
			out += "yes"
			util.ExpectEqual("bos.go copyUpload II", i+1, t.Errorf, out,
				testBosHandler.utilUploadFileArgVal)
		}
	}
	testBosCli.bosClient = tempClientBos
}

type copyType struct {
	srcPath        string
	dstPath        string
	dstBucketName  string
	finalObjectKey string
	storageClass   string
	downLoadTmp    string
	recursive      bool
}

func TestCopy(t *testing.T) {
	pathPrefix := "test_copy"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go copy %s", err.Error())
		}
	}()

	testBosCli.bosClient = tempFakeBosClient
	testCases := []copyType{
		copyType{
			srcPath:   pathPrefix,
			dstPath:   "bos:/liupeng-bj",
			recursive: true,
		},
		copyType{
			srcPath:   pathPrefix,
			dstPath:   "bos://liupeng-bj",
			recursive: true,
		},
		copyType{
			srcPath:   "bos:/0/",
			dstPath:   "./",
			recursive: true,
		},
		copyType{
			srcPath:   "bos:/cli-test/process/",
			dstPath:   "bos:/bucket/",
			recursive: true,
		},
	}
	for _, tCase := range testCases {
		testBosCli.Copy(tCase.srcPath, tCase.dstPath, tCase.storageClass, tCase.downLoadTmp,
			tCase.recursive, true, true, true, false)
	}
}

type syncPreProcessType struct {
	srcPath              string
	dstPath              string
	storageClass         string
	exclude              []string
	include              []string
	excludeTime          []string
	includeTime          []string
	concurrency          int
	del                  bool
	yes                  bool
	srcBucketName        string
	srcObjectKey         string
	dstBucketName        string
	dstObjectKey         string
	srcType              string
	dstType              string
	syncType             string
	syncProcessingNum    int
	multiUploadThreadNum int64
	confirmContent       string
	code                 BosCliErrorCode
	isSuc                bool
	filterDelConfirm     string
}

func TestSyncPreProcess(t *testing.T) {
	pathPrefix := "./test_sync_pre"
	if err := initListFileCases(pathPrefix, true); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go sync pre %s", err.Error())
		}
	}()

	defaultProceNum, ok := bceconf.ServerConfigProvider.GetSyncProcessingNum()
	if !ok {
		t.Errorf("sync pre GetMultiUploadThreadNum failed")
		return
	}
	defaultThreadNum, ok := bceconf.ServerConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		t.Errorf("sync pre GetMultiUploadThreadNum failed")
		return
	}

	testCases := []syncPreProcessType{
		//1
		syncPreProcessType{
			excludeTime: []string{"13213"},
			includeTime: []string{"1132123"},
			code:        BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG,
			isSuc:       false,
		},
		//2
		syncPreProcessType{
			srcPath: "./bucket",
			code:    boscmd.LOCAL_PATH_NOT_EXIST,
			isSuc:   false,
		},
		//3
		syncPreProcessType{
			srcPath: "./bos_test.go",
			code:    BOSCLI_SYNC_UPLOAD_SRC_MUST_DIR,
			isSuc:   false,
		},
		//4
		syncPreProcessType{
			srcPath: "bos:/bucket",
			dstPath: "./bos_test.go",
			code:    BOSCLI_SYNC_DOWN_DST_MUST_DIR,
			isSuc:   false,
		},
		//5
		syncPreProcessType{
			srcPath: pathPrefix,
			dstPath: pathPrefix,
			code:    BOSCLI_SYNC_LOCAL_TO_LOCAL,
			isSuc:   false,
		},
		//6
		syncPreProcessType{
			srcPath:     pathPrefix,
			dstPath:     "bos:/list_file",
			concurrency: -1,
			code:        BOSCLI_SYNC_PROCESS_NUM_LESS_ZERO,
			isSuc:       false,
		},
		//7
		syncPreProcessType{
			srcPath:              pathPrefix,
			dstPath:              "bos:/list_file",
			concurrency:          0,
			dstBucketName:        "list_file",
			srcType:              IS_LOCAL,
			dstType:              IS_BOS,
			syncType:             IS_LOCAL + IS_BOS,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//8
		syncPreProcessType{
			srcPath:              pathPrefix,
			dstPath:              "bos:/list_file/testkey/",
			concurrency:          0,
			dstBucketName:        "list_file",
			dstObjectKey:         "testkey/",
			srcType:              IS_LOCAL,
			dstType:              IS_BOS,
			syncType:             IS_LOCAL + IS_BOS,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//9
		syncPreProcessType{
			srcPath:              pathPrefix,
			dstPath:              "bos:/list_file/testkey/",
			concurrency:          6,
			dstBucketName:        "list_file",
			dstObjectKey:         "testkey/",
			srcType:              IS_LOCAL,
			dstType:              IS_BOS,
			syncType:             IS_LOCAL + IS_BOS,
			syncProcessingNum:    6,
			multiUploadThreadNum: defaultThreadNum,
			del:                  true,
			yes:                  true,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//10
		syncPreProcessType{
			srcPath:              pathPrefix,
			dstPath:              "bos:/list_file/testkey/",
			concurrency:          6,
			dstBucketName:        "list_file",
			dstObjectKey:         "testkey/",
			srcType:              IS_LOCAL,
			dstType:              IS_BOS,
			syncType:             IS_LOCAL + IS_BOS,
			syncProcessingNum:    6,
			multiUploadThreadNum: defaultThreadNum,
			del:                  true,
			confirmContent:       "yes",
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//11
		syncPreProcessType{
			srcPath:        pathPrefix,
			dstPath:        "bos:/list_file/testkey/",
			concurrency:    6,
			dstBucketName:  "list_file",
			dstObjectKey:   "testkey/",
			srcType:        IS_LOCAL,
			dstType:        IS_BOS,
			syncType:       IS_LOCAL + IS_BOS,
			del:            true,
			confirmContent: "no",
			code:           BOSCLI_OPRATION_CANCEL,
			isSuc:          false,
		},
		//12
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              "bos:/list_file/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			dstBucketName:        "list_file",
			dstObjectKey:         "testkey/",
			srcType:              IS_BOS,
			dstType:              IS_BOS,
			syncType:             IS_BOS + IS_BOS,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//13
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//14 both exclude and include are empty
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			exclude:              []string{},
			include:              []string{},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//15 one of exclude and include is empty
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			exclude:              []string{},
			include:              []string{"test"},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//16 both exclude and include exist
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			exclude:              []string{"test"},
			include:              []string{"test"},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_SYNC_EXCLUDE_INCLUDE_TOG,
			isSuc:                false,
		},
		//17 both excludeTime and includeTime exist
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			include:              []string{"test"},
			excludeTime:          []string{"time"},
			includeTime:          []string{"test"},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_SYNC_EXCLUDE_INCLUDE_TIME_TOG,
			isSuc:                false,
		},
		//18 both excludeTime and includeTime exist
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			include:              []string{"test"},
			excludeTime:          []string{},
			includeTime:          []string{"yes"},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			isSuc:                true,
		},
		//19 both filter and del exist: yes
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			include:              []string{"test"},
			excludeTime:          []string{},
			includeTime:          []string{},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OK,
			filterDelConfirm:     "yes",
			isSuc:                true,
		},
		//20 both filter and del exist: no
		syncPreProcessType{
			srcPath:              "bos:/list_file",
			dstPath:              pathPrefix + "/testkey/",
			concurrency:          0,
			srcBucketName:        "list_file",
			include:              []string{"test"},
			excludeTime:          []string{},
			includeTime:          []string{},
			srcType:              IS_BOS,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    defaultProceNum,
			multiUploadThreadNum: defaultThreadNum,
			code:                 BOSCLI_OPRATION_CANCEL,
			del:                  true,
			filterDelConfirm:     "no",
			isSuc:                false,
		},
	}
	tempStdin := os.Stdin
	for i, tCase := range testCases {
		var (
			fd           *os.File
			err          error
			tempFileName string
		)

		if tCase.filterDelConfirm != "" {
			t.Logf("id %d: confirm: %s\n", i+1, tCase.filterDelConfirm)
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n",
				tCase.filterDelConfirm)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		} else if tCase.del && !tCase.yes {
			if tCase.confirmContent == "" {
				t.Errorf("sync pre id %d don't set confirmContent", i+1)
				continue
			}
			fd, tempFileName, err = util.CreateAnRandomFileWithContent("%s\n", tCase.confirmContent)
			if err != nil {
				t.Errorf("create stdin input file filed")
				continue
			}
			defer os.Remove(tempFileName)
			os.Stdin = fd
		}

		args, code, err := testBosCli.syncPreProcess(tCase.srcPath, tCase.dstPath,
			tCase.storageClass, tCase.exclude, tCase.include, tCase.excludeTime, tCase.includeTime,
			tCase.concurrency, tCase.del, tCase.yes)

		util.ExpectEqual("bos.go sync pre I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("bos.go sync pre II", i+1, t.Errorf, tCase.code, code)
		if tCase.isSuc {
			util.ExpectEqual("bos.go sync pre III", i+1, t.Errorf, tCase.srcType, args.srcType)
			util.ExpectEqual("bos.go sync pre IV", i+1, t.Errorf, tCase.dstType, args.dstType)
			util.ExpectEqual("bos.go sync pre V", i+1, t.Errorf, tCase.syncType, args.syncType)
			util.ExpectEqual("bos.go sync pre VI", i+1, t.Errorf, tCase.syncProcessingNum,
				args.syncProcessingNum)
			util.ExpectEqual("bos.go sync pre VII", i+1, t.Errorf, tCase.multiUploadThreadNum,
				args.multiUploadThreadNum)
			util.ExpectEqual("bos.go sync pre VIII", i+1, t.Errorf, tCase.srcObjectKey,
				args.srcObjectKey)
			util.ExpectEqual("bos.go sync pre IX", i+1, t.Errorf, tCase.srcBucketName,
				args.srcBucketName)
			util.ExpectEqual("bos.go sync pre X", i+1, t.Errorf, tCase.dstBucketName,
				args.dstBucketName)
			util.ExpectEqual("bos.go sync pre XI", i+1, t.Errorf, tCase.dstObjectKey,
				args.dstObjectKey)
		}
	}
	os.Stdin = tempStdin
}

type syncExecuteType struct {
	srcPath              string
	dstPath              string
	srcBucketName        string
	srcObjectKey         string
	dstBucketName        string
	dstObjectKey         string
	downLoadTmp          string
	srcType              string
	dstType              string
	syncType             string // IS_LOCAL or IS_BOS
	syncKind             string // time-size, time-size-crc32, only-crc32
	syncProcessingNum    int
	multiUploadThreadNum int64

	exclude       []string
	include       []string
	excludeTime   []string
	includeTime   []string
	excludeDelete []string

	storageClass string
	del          bool
	yes          bool
	dryrun       bool

	failed    int
	successed int
	code      BosCliErrorCode
	isSuc     bool
}

func TestSyncExecute(t *testing.T) {
	var (
		fRetCode BosCliErrorCode
		fErr     error
	)

	pathPrefix := "./test_sync_exe"
	if err := initListFileCases(pathPrefix, false); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go sync exe %s", err.Error())
		}
	}()

	testCases := []syncExecuteType{
		//1
		syncExecuteType{
			srcPath: "root/~~~/`````",
			srcType: IS_LOCAL,
			code:    BOSCLI_EMPTY_CODE,
			isSuc:   false,
		},
		//2
		syncExecuteType{
			srcPath:       "bos:/error",
			srcType:       IS_BOS,
			srcBucketName: "error",
			code:          BOSCLI_EMPTY_CODE,
			isSuc:         false,
		},
		//3
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              pathPrefix,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            8,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//4
		syncExecuteType{
			srcPath:              pathPrefix,
			srcType:              IS_LOCAL,
			dstPath:              "bos:/0/a/",
			dstBucketName:        "0",
			dstObjectKey:         "a/",
			dstType:              IS_BOS,
			syncType:             IS_LOCAL + IS_BOS,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            17,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//5
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "bos:/0/a/",
			dstBucketName:        "0",
			dstObjectKey:         "a/",
			dstType:              IS_BOS,
			syncType:             IS_BOS + IS_BOS,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            0,
			failed:               0,
			del:                  true,
			dryrun:               true,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//6
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "bos:/0/a/",
			dstBucketName:        "0",
			dstObjectKey:         "a/",
			dstType:              IS_BOS,
			syncType:             IS_BOS + IS_BOS,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            8,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//7
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "bos:/0/a/",
			dstBucketName:        "0",
			dstObjectKey:         "a/",
			dstType:              IS_BOS,
			syncType:             IS_BOS + IS_BOS,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            16,
			failed:               0,
			del:                  true,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//8
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              pathPrefix,
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            25,
			failed:               0,
			del:                  true,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//9
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "./xxx/list_file",
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            8,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//10 with exclude
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "./xxx/list_file",
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			exclude:              []string{"bos:/cli-test/progress/*"},
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            0,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
		//11
		syncExecuteType{
			srcPath:              "bos:/cli-test/progress/",
			srcType:              IS_BOS,
			srcBucketName:        "cli-test",
			srcObjectKey:         "progress/",
			dstPath:              "./xxx/list_file",
			dstType:              IS_LOCAL,
			syncType:             IS_BOS + IS_LOCAL,
			include:              []string{"bos:/cli-test/progress/*"},
			syncProcessingNum:    5,
			multiUploadThreadNum: 5,
			successed:            8,
			failed:               0,
			code:                 BOSCLI_EMPTY_CODE,
			isSuc:                true,
		},
	}
	testBosCli.bosClient = tempFakeBosClient

	for i, tCase := range testCases {
		var (
			filter       *bosFilter = nil
			deleteFilter *bosFilter = nil
		)
		fmt.Printf("start id: %d\n", i+1)

		args := &syncArgs{
			srcPath:              tCase.srcPath,
			dstPath:              tCase.dstPath,
			srcBucketName:        tCase.srcBucketName,
			srcObjectKey:         tCase.srcObjectKey,
			dstBucketName:        tCase.dstBucketName,
			dstObjectKey:         tCase.dstObjectKey,
			srcType:              tCase.srcType,
			dstType:              tCase.dstType,
			syncType:             tCase.syncType,
			syncProcessingNum:    tCase.syncProcessingNum,
			multiUploadThreadNum: tCase.multiUploadThreadNum,
		}
		if len(tCase.exclude) > 0 || len(tCase.include) > 0 || len(tCase.includeTime) > 0 ||
			len(tCase.excludeTime) > 0 {
			filter, fRetCode, fErr = newSyncFilter(tCase.exclude, tCase.include, tCase.excludeTime,
				tCase.includeTime, args.srcType == IS_LOCAL)
			if fRetCode != BOSCLI_OK {
				t.Errorf("ID: %d, get new filter error : %s, code: %s", i+1, fErr, fRetCode)
			}
		}

		if len(tCase.excludeDelete) > 0 {
			deleteFilter, fRetCode, fErr = newSyncFilter(tCase.excludeDelete, []string{}, []string{},
				[]string{}, args.srcType == IS_LOCAL)
			if fRetCode != BOSCLI_OK {
				t.Errorf("ID: %d, get new delete filter error : %s, code: %s", i+1, fErr, fRetCode)
			}
		}

		result, code, err := testBosCli.syncExecute(filter, deleteFilter, args, tCase.storageClass,
			tCase.downLoadTmp, tCase.syncKind, tCase.del, tCase.dryrun, true)

		if tCase.isSuc != (err == nil) {
			t.Errorf("get error : %s", err)
			util.ExpectEqual("bos.go sync exe I", i+1, t.Errorf, tCase.isSuc, err == nil)
		}
		util.ExpectEqual("bos.go sync exe II", i+1, t.Errorf, tCase.code, code)
		if tCase.isSuc {
			t.Logf("id %d sucessed %d failed %d", i+1, result.successed, result.failed)
			if err != nil {
				t.Errorf("get error : %s", err)
				continue
			}
			util.ExpectEqual("bos.go sync exe III", i+1, t.Errorf, tCase.failed, result.failed)
			util.ExpectEqual("bos.go sync exe IV", i+1, t.Errorf, tCase.successed, result.successed)
		}
	}
}

type syncType struct {
	srcPath       string
	dstPath       string
	storageClass  string
	downLoadTmp   string
	syncType      string
	exclude       []string
	include       []string
	excludeTime   []string
	includeTime   []string
	excludeDelete []string
	concurrency   int
	del           bool
	yes           bool
	dryrun        bool
	quiet         bool
	disableBar    bool
	restart       bool
}

func TestSync(t *testing.T) {
	pathPrefix := "test_sync"
	if err := initListFileCases(pathPrefix, false); err != nil {
		t.Errorf("TestCopyUploadExecute create file test dir and file failed")
		return
	}
	defer func() {
		if err := removeListFileCases(pathPrefix); err != nil {
			t.Errorf("bos.go sync %s", err.Error())
		}
	}()

	testCases := []syncType{
		syncType{
			srcPath:     pathPrefix,
			dstPath:     "bos:/0/key",
			concurrency: 0,
		},
	}

	for _, tCase := range testCases {
		testBosCli.Sync(tCase.srcPath, tCase.dstPath, tCase.storageClass, tCase.downLoadTmp,
			tCase.syncType, tCase.exclude, tCase.include, tCase.excludeTime, tCase.includeTime,
			tCase.excludeDelete, tCase.concurrency, tCase.del, tCase.dryrun, tCase.yes, tCase.quiet,
			tCase.disableBar, tCase.restart)
	}
}
