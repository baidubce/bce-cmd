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
	"path"
	"strings"
	"time"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"utils/util"
)

var (
	DEFAULT_TMEP_FILE_SIZE = 1 << 20 // 1MB
	COMMON_FILE            = 1
	APPEND_FILE            = 2
	MULTI_PART_FILE        = 3
	UPLOAD_CHECK_PROPMT    = "上传"
)

type UploadCheck struct {
	ak         string
	sk         string
	endpoint   string
	bucketName string
	objectName string
	localPath  string
}

func (u *UploadCheck) getCheckType() string {
	return UPLOAD_CHECK_PROPMT
}

func (u *UploadCheck) requestInit(req ProbeRequest) BosProbeErrorCode {
	uploadResp, ok := req.(*UploadCheckRes)
	if !ok {
		return PROBE_INIT_REQUEST_FAILED
	}
	u.ak = uploadResp.Ak
	u.sk = uploadResp.Sk
	u.endpoint = uploadResp.Endpoint
	u.bucketName = uploadResp.BucketName
	u.objectName = uploadResp.ObjectName
	u.localPath = uploadResp.LocalPath
	return CODE_SUCCESS
}

func (u *UploadCheck) requestCheck() BosProbeErrorCode {
	if u.bucketName == "" {
		return LOCAL_ARGS_NO_BUCKET
	}
	return CODE_SUCCESS
}

func (u *UploadCheck) getEndpoint(bosClient *bos.Client) (string, BosProbeErrorCode, error) {
	if u.endpoint != "" {
		return u.endpoint, CODE_SUCCESS, nil
	}
	retEndPiont, err := boscmd.GetEndpointOfBucket(bosClient, u.bucketName)
	if err != nil {
		return "", LOCAL_GET_ENDPOINT_BUCKET_FAILED, err
	}
	u.endpoint = retEndPiont
	return retEndPiont, CODE_SUCCESS, nil
}

func (u *UploadCheck) uploadFile(sdk goSdk, bucketName, objectName, localPath, content string,
	mod int) (
	*serviceResp, error) {
	var (
		uploadResp *serviceResp
		uploadErr  error
	)
	switch mod {
	case COMMON_FILE:
		if localPath == "" {
			uploadResp, uploadErr = sdk.putObjectFromString(bucketName, objectName, content)
		} else {
			uploadResp, uploadErr = sdk.putObjectFromFile(bucketName, objectName, localPath)
		}
	default:
		return nil, fmt.Errorf("unsppourt model")
	}
	return uploadResp, uploadErr
}

func (u *UploadCheck) checkImpl(sdk goSdk) (*serviceResp, *objectDetail, *resultStatic) {
	var (
		err                error
		mod                = 1 // 1 common; 2 append; 3 multi-part
		fileSize           int64
		useTimeMillisecond int64
		checkResult        = &resultStatic{}
		fileContent        string
	)

	//generate temp file for upload and get file size
	if u.localPath == "" {
		fileSize = int64(DEFAULT_TMEP_FILE_SIZE)
		fileContent = util.GetRandomString(fileSize)
	} else {
		if ok := util.DoesFileExist(u.localPath); !ok {
			checkResult.code = boscmd.LOCAL_FILE_NOT_EXIST
			return nil, nil, checkResult
		}
		fileSize, err = util.GetSizeOfFile(u.localPath)
		if err != nil {
			checkResult.code = boscmd.LOCAL_OPEN_FILE_FAILED
			checkResult.err = err
			return nil, nil, checkResult
		}
	}

	//generate object name
	if u.objectName == "" || strings.HasSuffix(u.objectName, boscmd.BOS_PATH_SEPARATOR) {
		if u.localPath != "" {
			u.objectName += path.Base(u.localPath)
		} else {
			u.objectName += generateRandomFileName()
		}
	}

	//start upload test
	start := time.Now().UnixNano()
	uploadResp, uploadErr := u.uploadFile(sdk, u.bucketName, u.objectName, u.localPath, fileContent,
		mod)
	end := time.Now().UnixNano()
	useTimeMillisecond = (end - start) / 1e6
	if uploadErr != nil {
		if severErr, ok := uploadErr.(*bce.BceServiceError); ok {
			checkResult.code = BosProbeErrorCode(severErr.Code)
		} else {
			checkResult.code = boscmd.LOCAL_BCECLIENTERROR
		}
		checkResult.err = uploadErr
		return uploadResp, nil, checkResult
	} else {
		objectInfo := &objectDetail{
			objectName:         u.objectName,
			objectSize:         fileSize,
			useTimeMillisecond: useTimeMillisecond,
		}
		checkResult.code = CODE_SUCCESS
		return uploadResp, objectInfo, checkResult
	}
}
