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
	"strconv"
	"strings"
	"testing"
	"time"
)

import (
	// 	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"utils/util"
)

var (
	handler = &cliHandler{}
)

func init() {
	if err := initConfig(); err != nil {
		os.Exit(1)
	}
}

type fakeBosClient struct {
	results    []*api.ListObjectsResult
	objectMeta *api.GetObjectMetaResult
}

func (b *fakeBosClient) HeadBucket(bucket string) error {
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
func (b *fakeBosClient) ListBuckets() (*api.ListBucketsResult, error) {
	return nil, fmt.Errorf("test")
}

// Fake PutBucket - create a new bucket
func (b *fakeBosClient) PutBucket(bucket string) (string, error) {
	return "", fmt.Errorf("test")
}

// Fake GetBucketLocation - get the location fo the given bucket
func (b *fakeBosClient) GetBucketLocation(bucket string) (string, error) {
	return "", fmt.Errorf("test")
}

// Fake ListObjects - list all objects of the given bucket
func (b *fakeBosClient) ListObjects(bucket string, args *api.ListObjectsArgs) (
	*api.ListObjectsResult, error) {
	var (
		marker int
		err    error
	)
	if args.Marker == "" {
		marker, err = strconv.Atoi(bucket)
	} else {
		marker, err = strconv.Atoi(args.Marker)
	}
	if err != nil {
		return nil, err
	}
	if marker < len(b.results) {
		return b.results[marker], nil
	}
	return nil, fmt.Errorf("Error in list objects")
}

// Fake DeleteBucket - delete a empty bucket
func (b *fakeBosClient) DeleteBucket(bucket string) error {

	return fmt.Errorf("test")
}

// Fake BasicGeneratePresignedUrl  generate an authorization url with expire time
func (b *fakeBosClient) BasicGeneratePresignedUrl(bucket string, object string,
	expireInSeconds int) string {
	return ""
}

// Fake DeleteMultipleObjectsFromKeyList - delete a list of objects with given key string array
func (b *fakeBosClient) DeleteMultipleObjectsFromKeyList(bucket string,
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
func (b *fakeBosClient) DeleteObject(bucket, object string) error {
	if bucket == "error" {
		return fmt.Errorf(bucket + object)
	}
	return nil
}

// Fake GetObjectMeta
func (b *fakeBosClient) GetObjectMeta(bucket, object string) (*api.GetObjectMetaResult, error) {
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
func (b *fakeBosClient) CopyObject(bucket, object, srcBucket, srcObject string,
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
func (b *fakeBosClient) BasicGetObjectToFile(bucket, object, localPath string) error {
	if bucket == "success" && object == "a/b/c" {
		return nil
	}
	return fmt.Errorf(bucket + object + localPath)
}

// Fake of PutObjectFromFile
func (b *fakeBosClient) PutObjectFromFile(bucket, object, fileName string,
	args *api.PutObjectArgs) (string, error) {
	if fileName == "success" {
		return "", nil
	}
	return "", fmt.Errorf("smail" + fileName + bucket + object + args.StorageClass)
}

// Fake of UploadSuperFile
func (b *fakeBosClient) UploadSuperFile(bucket, object, fileName, storageClass string) error {
	if fileName == "success" {
		return nil
	}
	return fmt.Errorf("big" + fileName + bucket + object + storageClass)
}

func (b *fakeBosClient) BasicUploadPart(bucket, object, uploadId string, partNumber int,
	content *bce.Body) (string, error) {
	return "", fmt.Errorf("usupport BasicUploadPart")
}

func (b *fakeBosClient) UploadPartFromBytes(bucket, object, uploadId string, partNumber int,
	content []byte, args *api.UploadPartArgs) (string, error) {
	return "", fmt.Errorf("Not support")
}

// Fake of PutBucketLifecycle
func (b *fakeBosClient) PutBucketLifecycleFromString(bucket, lifecycle string) error {
	return nil
}

// Fake of GetBucketLifecycle
func (b *fakeBosClient) GetBucketLifecycle(bucket string) (*api.GetBucketLifecycleResult, error) {
	return nil, nil
}

// Fake of DeleteBucketLifecycle
func (b *fakeBosClient) DeleteBucketLifecycle(bucket string) error {
	return nil
}

// Fake of PutBucketLoggingFromStruct
func (b *fakeBosClient) PutBucketLoggingFromStruct(bucket string,
	obj *api.PutBucketLoggingArgs) error {

	return nil
}

// Fake of GetBucketLogging
func (b *fakeBosClient) GetBucketLogging(bucket string) (*api.GetBucketLoggingResult, error) {
	return nil, nil
}

// Fake of DeleteBucketLogging
func (b *fakeBosClient) DeleteBucketLogging(bucket string) error {
	return nil
}

// Fake of PutBucketStorageclass
func (b *fakeBosClient) PutBucketStorageclass(bucket, storageClass string) error {
	return nil
}

// Fake of GetBucketStorageclass
func (b *fakeBosClient) GetBucketStorageclass(bucket string) (string, error) {
	return "", nil
}

// Fake of PutBucketAclFromCanned
func (b *fakeBosClient) PutBucketAclFromCanned(bucket, cannedAcl string) error {
	return nil
}

// Fake of PutBucketAcl
func (b *fakeBosClient) PutBucketAclFromString(bucket, acl string) error {
	return nil
}

// Fake of GetBucketAcl
func (b *fakeBosClient) GetBucketAcl(bucket string) (*api.GetBucketAclResult, error) {
	return nil, nil
}

// Fake of GetBucketStorageclass
func (b *fakeBosClient) UploadPartCopy(bucket, object, srcBucket, srcObject, uploadId string,
	partNumber int, args *api.UploadPartCopyArgs) (*api.CopyObjectResult, error) {
	return nil, fmt.Errorf("Not support")
}

// Fake of GetBucketStorageclass
func (b *fakeBosClient) InitiateMultipartUpload(bucket, object, contentType string,
	args *api.InitiateMultipartUploadArgs) (*api.InitiateMultipartUploadResult, error) {

	return nil, fmt.Errorf("Not support")
}

func (b *fakeBosClient) AbortMultipartUpload(bucket, object, uploadId string) error {
	return fmt.Errorf("Not support")
}

func (b *fakeBosClient) CompleteMultipartUploadFromStruct(bucket, object, uploadId string,
	parts *api.CompleteMultipartUploadArgs,
	meta map[string]string) (*api.CompleteMultipartUploadResult, error) {

	return nil, fmt.Errorf("Not support")
}

func (b *fakeBosClient) GetObject(bucket, object string, responseHeaders map[string]string,
	ranges ...int64) (*api.GetObjectResult, error) {

	return nil, fmt.Errorf("Not support")
}

type listObjectIteratorType struct {
	bucketName string
	objectKey  string
	bosClient  bosClientInterface
	output     []*listFileResult
	isDir      bool

	// filter
	exclude     []string
	include     []string
	excludeTime []string
	includeTime []string
}

func TestListObjectIterator(t *testing.T) {
	testCases := []listObjectIteratorType{
		// list dir
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "",
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
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
						IsTruncated: false,
					},
				},
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					file: &fileDetail{
						key:   "a/b",
						size:  100,
						mtime: 1136214245,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "a/c",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:          "a/d",
						size:         300,
						mtime:        1509635045,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					ended: true,
				},
			},
		},
		// list dir filter: exclude
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/",
			exclude:    []string{"bos:/0/a/*"},
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
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
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					ended: true,
				},
			},
		},
		// list dir: two include
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/",
			include:    []string{"bos:/0/a/c", "bos:/0/a/g", "bos:/0/a/h"},
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
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
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					file: &fileDetail{
						key:   "c",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "g",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:          "h",
						size:         300,
						mtime:        1509635045,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					ended: true,
				},
			},
		},

		// list dir: two exclude
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/",
			exclude:    []string{"bos:/0/a/b", "bos:/0/a/d", "bos:/0/a/f"},
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
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
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					file: &fileDetail{
						key:   "c",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "g",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:          "h",
						size:         300,
						mtime:        1509635045,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					ended: true,
				},
			},
		},

		// list dir
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/",
			bosClient: &fakeBosClient{
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
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					dir: &dirDetail{
						key: "dir/",
					},
					isDir: true,
				},
				&listFileResult{
					dir: &dirDetail{
						key: "dir2/",
					},
					isDir: true,
				},
				&listFileResult{
					file: &fileDetail{
						key:   "b",
						size:  100,
						mtime: 1136214245,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "c",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:          "d",
						size:         300,
						mtime:        1509635045,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					dir: &dirDetail{
						key: "eir/",
					},
					isDir: true,
				},
				&listFileResult{
					dir: &dirDetail{
						key: "eir2/",
					},
					isDir: true,
				},
				&listFileResult{
					file: &fileDetail{
						key:   "f",
						size:  100,
						mtime: 1136214245,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "g",
						size:  200,
						mtime: 1478099045,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:          "h",
						size:         300,
						mtime:        1509635045,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					ended: true,
				},
			},
		},
		// list dir time error
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "",
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
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
								LastModified: "-11-02T15:04:05Z",
								Size:         300,
								StorageClass: "STANDARD",
							},
						},
						IsTruncated: false,
					},
				},
			},
			isDir: true,
			output: []*listFileResult{
				&listFileResult{
					file: &fileDetail{
						key:   "a/b",
						size:  100,
						mtime: 1136214245,
					},
				},
				&listFileResult{
					file: &fileDetail{
						key:   "a/c",
						size:  200,
						mtime: 1478099045,
					},
				},
				nil,
			},
		},
		// single object
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/b/c",
			bosClient: &fakeBosClient{
				objectMeta: &api.GetObjectMetaResult{
					api.ObjectMeta{
						LastModified:  "Wed, 06 Apr 2016 06:34:40 GMT",
						ContentLength: 100,
						StorageClass:  "STANDARD",
					},
				},
			},
			isDir: false,
			output: []*listFileResult{
				&listFileResult{
					file: &fileDetail{
						path:         "a/b/c",
						key:          "c",
						size:         100,
						mtime:        1459924480,
						storageClass: "STANDARD",
					},
				},
				&listFileResult{
					ended: true,
				},
			},
		},

		// single object time error
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/b/c",
			bosClient: &fakeBosClient{
				objectMeta: &api.GetObjectMetaResult{
					api.ObjectMeta{
						LastModified:  "06 Apr 2016 06:34:40 GMT",
						ContentLength: 100,
						StorageClass:  "STANDARD",
					},
				},
			},
			isDir:  false,
			output: []*listFileResult{nil},
		},

		// single object err
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "a/b/c",
			bosClient: &fakeBosClient{
				objectMeta: nil,
			},
			isDir:  false,
			output: []*listFileResult{nil},
		},
		// single object err 404
		listObjectIteratorType{
			bucketName: "0",
			objectKey:  "404",
			bosClient: &fakeBosClient{
				objectMeta: nil,
			},
			isDir:  false,
			output: []*listFileResult{nil},
		},
	}
	for i, tCase := range testCases {
		var (
			filter   *bosFilter = nil
			fRetCode BosCliErrorCode
			fErr     error
		)

		if len(tCase.exclude) > 0 || len(tCase.include) > 0 || len(tCase.includeTime) > 0 ||
			len(tCase.excludeTime) > 0 {
			filter, fRetCode, fErr = newSyncFilter(tCase.exclude, tCase.include, tCase.excludeTime,
				tCase.includeTime, false)
			if fRetCode != BOSCLI_OK {
				t.Errorf("ID: %d, get new filter error : %s, code: %s", i+1, fErr, fRetCode)
			}
		}

		newIter := NewObjectListIterator(tCase.bosClient, filter, tCase.bucketName, tCase.objectKey,
			"", true, true, tCase.isDir, false, 1000)
		index := 0
		for {
			ret, err := newIter.next()
			if tCase.output[index] == nil {
				util.ExpectEqual("handler.go listTest I", i+1, t.Errorf, true, err != nil)
				break
			} else if tCase.output[index].err != nil {
				util.ExpectEqual("handler.go listTest II", i+1, t.Errorf, true, ret.err != nil)
				break
			} else if tCase.output[index].ended == true {
				util.ExpectEqual("handler.go listTest II", i+1, t.Errorf, true, ret.ended)
				break
			} else {
				t.Logf("%d index %d", i+1, index)
				t.Logf("%d err %d", i+1, err)
				t.Logf("%d ret %v", i+1, ret)
				t.Logf("%d need %v", i+1, tCase.output[index])
				if tCase.output[index].isDir {
					util.ExpectEqual("handler.go listTest dir", i+1, t.Errorf,
						tCase.output[index].dir.key, ret.dir.key)
				} else {
					util.ExpectEqual("handler.go listTest III", i+1, t.Errorf,
						tCase.output[index].file.key, ret.file.key)
					util.ExpectEqual("handler.go listTest IV", i+1, t.Errorf,
						tCase.output[index].file.size, ret.file.size)
					util.ExpectEqual("handler.go listTest V", i+1, t.Errorf,
						tCase.output[index].file.mtime, ret.file.mtime)
					util.ExpectEqual("handler.go listTest VI", i+1, t.Errorf,
						tCase.output[index].file.storageClass, ret.file.storageClass)
				}
			}
			index++
		}
	}
}

type multiDeleteObjectsWithRetryType struct {
	bosClient  bosClientInterface
	bucketName string
	objectList []string
	output     []api.DeleteObjectResult
	isSuc      bool
}

func TestMultiDeleteObjectsWithRetry(t *testing.T) {
	testCases := []multiDeleteObjectsWithRetryType{
		multiDeleteObjectsWithRetryType{
			bosClient:  &fakeBosClient{},
			bucketName: "bucket1",
			objectList: []string{
				"object1",
				"object2",
				"object3",
				"object4",
			},
			output: nil,
			isSuc:  true,
		},
		multiDeleteObjectsWithRetryType{
			bosClient:  &fakeBosClient{},
			bucketName: "bucket1",
			objectList: []string{
				"object1",
				"TFAILED-2-502-object1",
				"TFAILED-2-502-object2",
				"TFAILED-2-503-object3",
				"TFAILED-2-504-object4",
				"object3",
				"object4",
			},
			output: nil,
			isSuc:  true,
		},
		multiDeleteObjectsWithRetryType{
			bosClient:  &fakeBosClient{},
			bucketName: "bucket1",
			objectList: []string{
				"object1",
				"TFAILED-2-502-object1",
				"TFAILED-2-502-object2",
				"TFAILED-2-503-object3",
				"TFAILED-2-504-object4",
				"TFAILED-3-504-object4",
				"TFAILED-3-505-object5",
				"TFAILED-3-NoSuchKey-object5",
				"object3",
				"object4",
			},
			output: []api.DeleteObjectResult{
				api.DeleteObjectResult{
					Key:  "TFAILED-3-NoSuchKey-object5retry",
					Code: "NoSuchKey",
				},
				api.DeleteObjectResult{
					Key:  "TFAILED-3-504-object4retryretry",
					Code: "504",
				},
				api.DeleteObjectResult{
					Key:  "TFAILED-3-505-object5retryretry",
					Code: "505",
				},
			},
			isSuc: false,
		},
	}
	for i, tCase := range testCases {
		ret, err := handler.multiDeleteObjectsWithRetry(tCase.bosClient, tCase.objectList,
			tCase.bucketName)
		t.Logf("%s", err)
		t.Logf("%v", ret)
		if tCase.isSuc {
			util.ExpectEqual("handler.go mDeleteRetry II", i+1, t.Errorf, true, ret == nil)
		} else {
			for j, val := range tCase.output {
				util.ExpectEqual("handler.go mDeleteRetry III", i+1, t.Errorf, val.Key,
					ret[j].Key)
				util.ExpectEqual("handler.go mDeleteRetry III", i+1, t.Errorf, val.Code,
					ret[j].Code)
			}
		}
	}
}

type multiDeleteDirType struct {
	bosClient  bosClientInterface
	bucketName string
	objectKey  string
	deleted    int
	isSuc      bool
}

func TestMultiDeleteDir(t *testing.T) {
	testCases := []multiDeleteDirType{
		multiDeleteDirType{
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
						Contents: []api.ObjectSummaryType{
							api.ObjectSummaryType{
								Key:          "object1",
								LastModified: "2006-01-02T15:04:05Z",
								Size:         100,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-502-object1",
								LastModified: "2016-11-02T15:04:05Z",
								Size:         200,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-502-object2",
								LastModified: "2017-11-02T15:04:05Z",
								Size:         300,
								StorageClass: "STANDARD",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-503-object3",
								LastModified: "2006-01-02T15:04:05Z",
								Size:         100,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-3-NoSuchKey-object5",
								LastModified: "2016-11-02T15:04:05Z",
								Size:         200,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-3-505-object5",
								LastModified: "2017-11-02T15:04:05Z",
								Size:         300,
								StorageClass: "STANDARD",
							},
						},
						IsTruncated: false,
					},
				},
			},
			bucketName: "0",
			objectKey:  "path/",
			deleted:    4,
			isSuc:      false,
		},
		multiDeleteDirType{
			bosClient: &fakeBosClient{
				results: []*api.ListObjectsResult{
					&api.ListObjectsResult{
						Contents: []api.ObjectSummaryType{
							api.ObjectSummaryType{
								Key:          "object1",
								LastModified: "2006-01-02T15:04:05Z",
								Size:         100,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-502-object1",
								LastModified: "2016-11-02T15:04:05Z",
								Size:         200,
								StorageClass: "",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-502-object2",
								LastModified: "2017-11-02T15:04:05Z",
								Size:         300,
								StorageClass: "STANDARD",
							},
							api.ObjectSummaryType{
								Key:          "TFAILED-2-503-object3",
								LastModified: "2006-01-02T15:04:05Z",
								Size:         100,
								StorageClass: "",
							},
						},
						IsTruncated: false,
					},
				},
			},
			bucketName: "0",
			objectKey:  "path/",
			deleted:    4,
			isSuc:      true,
		},
	}
	for i, tCase := range testCases {
		ret, _ := handler.multiDeleteDir(tCase.bosClient, tCase.bucketName, tCase.objectKey)
		util.ExpectEqual("handler.go multiDeleteDir II", i+1, t.Errorf, tCase.deleted, ret)
	}
}

type utilDeleteObjectType struct {
	bucketName string
	objectKey  string
	isSuc      bool
}

func TestUtilDeleteObject(t *testing.T) {
	testCases := []utilDeleteObjectType{
		utilDeleteObjectType{
			bucketName: "bucket1",
			objectKey:  "bucket1",
			isSuc:      true,
		},
		utilDeleteObjectType{
			bucketName: "error",
			objectKey:  "bucket1",
			isSuc:      false,
		},
	}
	bosClient := &fakeBosClient{}
	for i, tCase := range testCases {
		ret := handler.utilDeleteObject(bosClient, tCase.bucketName, tCase.objectKey)
		if tCase.isSuc {
			util.ExpectEqual("handler.go utilDeleteObject I", i+1, t.Errorf,
				true, ret == nil)
		} else {
			util.ExpectEqual("handler.go utilDeleteObject I", i+1, t.Errorf,
				tCase.bucketName+tCase.objectKey, ret.Error())
		}
	}
}

type utilCopyObjectType struct {
	srcBucket    string
	srcObject    string
	dstBucket    string
	dstObject    string
	storageClass string
	fileSize     int64
	fileMtime    int64
	restart      bool
	isSuc        bool
	err          error
}

func TestUtilCopyObject(t *testing.T) {
	testCases := []utilCopyObjectType{
		utilCopyObjectType{
			srcBucket:    "sbucket",
			srcObject:    "sobject",
			dstBucket:    "dbucket",
			dstObject:    "dobject",
			storageClass: "",
			fileSize:     100,
			fileMtime:    time.Now().Unix(),
			isSuc:        true,
		},
		utilCopyObjectType{
			srcBucket:    "sbucket",
			srcObject:    "sobject",
			dstBucket:    "dbucket",
			dstObject:    "dobject",
			storageClass: "",
			fileSize:     100 << 30,
			fileMtime:    time.Now().Unix(),
			isSuc:        false,
			err:          fmt.Errorf("Not support"),
		},
		utilCopyObjectType{
			srcBucket:    "error",
			srcObject:    "sobject",
			dstBucket:    "dbucket",
			dstObject:    "dobject",
			storageClass: "",
			fileSize:     100,
			fileMtime:    time.Now().Unix(),
			isSuc:        false,
			err:          fmt.Errorf("error"),
		},
	}
	bosClient := &fakeBosClient{}
	srcBosClient := &fakeBosClient{}
	for i, tCase := range testCases {

		ret := handler.utilCopyObject(srcBosClient, bosClient, tCase.srcBucket, tCase.srcObject,
			tCase.dstBucket, tCase.dstObject, tCase.storageClass, tCase.fileSize, tCase.fileMtime,
			time.Now().Unix(), tCase.restart)

		if tCase.isSuc {
			util.ExpectEqual("handler.go utilCopyObject I", i+1, t.Errorf,
				true, ret == nil)
		} else {
			util.ExpectEqual("handler.go utilCopyObject I", i+1, t.Errorf,
				tCase.err, ret)
		}
	}
}

type utilDownloadObjectType struct {
	srcBucket           string
	srcObject           string
	localPath           string
	retPath             string
	downLoadTmp         string
	yes                 bool
	isSuc               bool
	err                 string
	fileSize            int64
	mtime               int64
	timeOfgetObjectInfo int64
	restart             bool
}

func TestUtilDownloadObject(t *testing.T) {
	absPath, err := util.Abs("./")
	if err != nil {
		t.Errorf("get absPath failed")
		return
	}

	now := time.Now().Unix()

	// 	if bucket == "success" && object == "a/b/c" && localPath == "./test.txt" {
	testCases := []utilDownloadObjectType{
		//1
		utilDownloadObjectType{
			srcBucket:           "success",
			srcObject:           "a/b/c",
			localPath:           "./test.txt",
			yes:                 true,
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./test.txt",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/test.txt", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		//3
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./test/file/",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/test/file/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		//5
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: ".",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		utilDownloadObjectType{
			srcBucket:           "bucket1",
			srcObject:           "a/b/c",
			localPath:           "/",
			yes:                 true,
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			err: "bucket1a/b/c" + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			isSuc: true,
		},
		//7
		utilDownloadObjectType{
			srcBucket:           "bucket1",
			srcObject:           "a/b/c",
			localPath:           "/root/",
			yes:                 true,
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
		utilDownloadObjectType{
			srcBucket:           "bucket1",
			srcObject:           "a/b/c",
			localPath:           "/root/cover.html",
			yes:                 true,
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
		//9
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: userHomeDir + "/debug_bos/file/",
			yes:       true,
			err: "bucket1a/b/c" + strings.Replace(userHomeDir+"/debug_bos/file/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               true,
		},
		//11
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./cover.out/",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./cover.out",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
		//13
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./test/test/test/",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/test/test/test/c", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
		utilDownloadObjectType{
			srcBucket: "bucket1",
			srcObject: "a/b/c",
			localPath: "./test/test/xtest/cover",
			yes:       true,
			err: "bucket1a/b/c" + absPath + strings.Replace("/test/test/xtest/cover", "/",
				util.OsPathSeparator, -1),
			fileSize:            1,
			mtime:               now,
			timeOfgetObjectInfo: now,
			isSuc:               false,
		},
	}
	os.RemoveAll("./test")
	os.Create("./cover.out")
	bosClient := &fakeBosClient{}
	for i, tCase := range testCases {
		ret := handler.utilDownloadObject(bosClient, tCase.srcBucket, tCase.srcObject,
			tCase.localPath, tCase.downLoadTmp, tCase.yes, tCase.fileSize, tCase.mtime,
			tCase.timeOfgetObjectInfo, tCase.restart)

		if !tCase.isSuc {
			util.ExpectEqual("handler.go utilDownloadObject I", i+1, t.Errorf, true, ret != nil)
			continue
		}
		if tCase.err == "" {
			util.ExpectEqual("handler.go utilDownloadObject II", i+1, t.Errorf, true, ret == nil)
			if ret != nil {
				t.Logf("unexpected %s", ret)
			}
			continue
		}
		if tCase.err != ret.Error() {
			util.ExpectEqual("handler.go utilDownloadObject III", i+1, t.Errorf, tCase.err,
				ret.Error())
		} else {
			t.Logf("%s", ret.Error())
		}
	}
	os.RemoveAll("./test")
}

type utilUploadFileType struct {
	srcPath             string
	relSrcPath          string
	dstBucket           string
	dstObject           string
	storageClass        string
	fileSize            int64
	fileMtime           int64
	timeOfgetObjectInfo int64
	restart             bool
	err                 string
}

func TestUtilUploadFile(t *testing.T) {
	testCases := []utilUploadFileType{
		utilUploadFileType{
			srcPath:             "sbucket",
			relSrcPath:          "sobject",
			dstBucket:           "dbucket",
			dstObject:           "dobject",
			storageClass:        "",
			fileSize:            100,
			fileMtime:           time.Now().Unix(),
			timeOfgetObjectInfo: time.Now().Unix(),
			err:                 "smailsobjectdbucketdobject",
		},
		utilUploadFileType{
			srcPath:             "success",
			relSrcPath:          "success",
			dstBucket:           "dbucket",
			dstObject:           "dobject",
			storageClass:        "",
			fileSize:            100,
			fileMtime:           time.Now().Unix(),
			timeOfgetObjectInfo: time.Now().Unix(),
		},
	}
	bosClient := &fakeBosClient{}
	for i, tCase := range testCases {
		ret := handler.utilUploadFile(bosClient, tCase.srcPath, tCase.relSrcPath, tCase.dstBucket,
			tCase.dstObject, tCase.storageClass, tCase.fileSize, tCase.fileMtime,
			tCase.timeOfgetObjectInfo, tCase.restart)
		if tCase.err == "" {
			util.ExpectEqual("handler.go utilUploadFile I", i+1, t.Errorf,
				true, ret == nil)
		} else {
			util.ExpectEqual("handler.go utilUploadFile I", i+1, t.Errorf,
				tCase.err, ret.Error())
		}
	}
}

type utilDeleteLocalFileType struct {
	srcPath string
	err     string
	build   bool
	isDir   bool
}

func TestUtilDeleteLocalFile(t *testing.T) {
	testCases := []utilDeleteLocalFileType{
		utilDeleteLocalFileType{
			srcPath: "./sbucket/test",
			build:   true,
			isDir:   false,
		},
		utilDeleteLocalFileType{
			srcPath: "./sbucket/test1/test",
			build:   true,
			isDir:   false,
		},
		utilDeleteLocalFileType{
			srcPath: "./sbucke/test2/test",
			err:     "smailsobjectdbucketdobject",
		},
		utilDeleteLocalFileType{
			srcPath: "./sbucket/test",
			isDir:   true,
			err:     "smailsobjectdbucketdobject",
		},
		utilDeleteLocalFileType{
			srcPath: "./sbucket/test",
			build:   true,
			isDir:   true,
			err:     "smailsobjectdbucketdobject",
		},
		utilDeleteLocalFileType{
			srcPath: "sbucket",
			build:   true,
		},
		utilDeleteLocalFileType{
			srcPath: "success",
			err:     "smailsobjectdbucketdobject",
		},
	}
	for i, tCase := range testCases {
		os.RemoveAll("./sbucket")
		if tCase.isDir && tCase.build {
			err := util.TryMkdir(tCase.srcPath)
			if err != nil {
				t.Errorf("make dir %s failed, %s", tCase.srcPath, err)
				continue
			}
		} else if tCase.build {
			absPath, err := util.Abs(tCase.srcPath)
			if err != nil {
				t.Errorf("get abs path of %s failed, %s", tCase.srcPath, err)
				continue
			}
			index := strings.LastIndex(absPath, util.OsPathSeparator)
			if index > 0 {
				err = util.TryMkdir(absPath[:index])
				if err != nil {
					t.Errorf("make dir %s failed, %s", absPath[:index], err)
					continue
				}
			}
			fd, err := os.Create(absPath)
			if err != nil {
				t.Errorf("create file %s failed, %s", absPath, err)
				continue
			}
			fmt.Fprintf(fd, "%s", absPath)
			fd.Close()
		}
		ret := handler.utilDeleteLocalFile(tCase.srcPath)
		util.ExpectEqual("handler.go utilDeleteLocalFile I", i+1, t.Errorf, tCase.err == "",
			ret == nil)
	}
}

type doesBucketExistType struct {
	srcPath string
	exist   bool
	isSuc   bool
}

func TestDoesBucketExist(t *testing.T) {
	testCases := []doesBucketExistType{
		doesBucketExistType{
			srcPath: "exist",
			exist:   true,
			isSuc:   true,
		},
		doesBucketExistType{
			srcPath: "notexist",
			exist:   false,
			isSuc:   true,
		},
		doesBucketExistType{
			srcPath: "error",
			exist:   false,
			isSuc:   false,
		},
		doesBucketExistType{
			srcPath: "forbidden",
			exist:   true,
			isSuc:   true,
		},
	}
	bosClient := &fakeBosClient{}
	for i, tCase := range testCases {
		ret, err := handler.doesBucketExist(bosClient, tCase.srcPath)
		util.ExpectEqual("handler.go doesBucketExist I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("handler.go doesBucketExist II", i+1, t.Errorf, tCase.exist, ret)
	}
}
