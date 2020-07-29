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

// .

package boscli

import (
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

// Interface for wrap go sdk
type bosClientInterface interface {
	HeadBucket(bucket string) error
	ListBuckets() (*api.ListBucketsResult, error)
	ListObjects(string, *api.ListObjectsArgs) (*api.ListObjectsResult, error)
	PutBucket(string) (string, error)
	DeleteBucket(string) error
	GetBucketLocation(string) (string, error)
	BasicGeneratePresignedUrl(string, string, int) string
	DeleteMultipleObjectsFromKeyList(string, []string) (*api.DeleteMultipleObjectsResult, error)
	DeleteObject(string, string) error
	GetObjectMeta(string, string) (*api.GetObjectMetaResult, error)
	CopyObject(string, string, string, string, *api.CopyObjectArgs) (
		*api.CopyObjectResult, error)
	BasicGetObjectToFile(string, string, string) error
	PutObjectFromFile(string, string, string, *api.PutObjectArgs) (string, error)
	UploadSuperFile(string, string, string, string) error
	PutBucketLifecycleFromString(string, string) error
	GetBucketLifecycle(string) (*api.GetBucketLifecycleResult, error)
	DeleteBucketLifecycle(string) error
	PutBucketLoggingFromStruct(string, *api.PutBucketLoggingArgs) error
	GetBucketLogging(string) (*api.GetBucketLoggingResult, error)
	DeleteBucketLogging(string) error
	PutBucketStorageclass(string, string) error
	GetBucketStorageclass(string) (string, error)
	PutBucketAclFromCanned(string, string) error
	PutBucketAclFromString(string, string) error
	GetBucketAcl(string) (*api.GetBucketAclResult, error)
	UploadPartCopy(string, string, string, string, string,
		int, *api.UploadPartCopyArgs) (*api.CopyObjectResult, error)
	BasicUploadPart(string, string, string, int, *bce.Body) (string, error)
	UploadPartFromBytes(string, string, string, int, []byte, *api.UploadPartArgs) (string, error)
	InitiateMultipartUpload(string, string, string, *api.InitiateMultipartUploadArgs) (
		*api.InitiateMultipartUploadResult, error)
	AbortMultipartUpload(bucket, object, uploadId string) error
	CompleteMultipartUploadFromStruct(string, string, string, *api.CompleteMultipartUploadArgs,
	) (*api.CompleteMultipartUploadResult, error)
	GetObject(string, string, map[string]string, ...int64) (*api.GetObjectResult, error)
}

// Interface for bos cli handler
type handlerInterface interface {
	multiDeleteDir(bosClientInterface, string, string) (int, error)
	multiDeleteObjectsWithRetry(bosClientInterface, []string, string) ([]api.DeleteObjectResult,
		error)
	utilDeleteObject(bosClientInterface, string, string) error
	utilCopyObject(bosClientInterface, bosClientInterface, string, string, string, string, string,
		int64, int64, int64, bool) error
	utilDownloadObject(bosClientInterface, string, string, string, string, bool, int64, int64, int64,
		bool) error
	utilUploadFile(bosClientInterface, string, string, string, string, string, int64, int64,
		int64, bool) error
	utilDeleteLocalFile(string) error
	doesBucketExist(bosClientInterface, string) (bool, error)
	CopySuperFile(bosClientInterface, bosClientInterface, string, string, string, string,
		string, int64, int64, int64, bool, string) error
}

// File Information: be used by BOS object and local file.
type fileDetail struct {
	path         string // both, full path (path of bos object don't contail bucket name)
	name         string // without replcae os sep to bos sep
	key          string // both
	realPath     string // local file, real path of symbolic link
	storageClass string // bos object
	crc32        string
	size         int64 // both
	mtime        int64 // both, last Modified time
	gtime        int64 // both, the time of get info of this object
	isDir        bool
	err          error // both
}

// Directory Information: be used by BOS pre.
type dirDetail struct {
	path string
	key  string
}

// End Information: be used by BOS.
type listEndInfo struct {
	nextMarker  string
	isTruncated bool
}

type listFileResult struct {
	file    *fileDetail
	dir     *dirDetail
	endInfo *listEndInfo
	err     error //have error when list object?
	isDir   bool  //is file or dir?
	ended   bool  //have get all object?
}

type fileListIterator interface {
	next() (*listFileResult, error)
}

type executeResult struct {
	failed    int
	successed int
}
