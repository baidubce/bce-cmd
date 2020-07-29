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
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"utils/util"
)

const (
	DOWNLOAD_CHECK_PROPMT = "下载"
)

type DownloadCheck struct {
	fileContent string
	ak          string
	sk          string
	fileUrl     string
	endpoint    string
	bucketName  string
	objectName  string
	localPath   string
}

func (d *DownloadCheck) requestInit(req ProbeRequest) BosProbeErrorCode {
	downloadResp, ok := req.(*DownloadCheckRes)
	if !ok {
		return PROBE_INIT_REQUEST_FAILED
	}
	d.ak = downloadResp.Ak
	d.sk = downloadResp.Sk
	d.fileUrl = downloadResp.FileUrl
	d.endpoint = downloadResp.Endpoint
	d.bucketName = downloadResp.BucketName
	d.objectName = downloadResp.ObjectName
	d.localPath = downloadResp.LocalPath
	return CODE_SUCCESS
}

func (d *DownloadCheck) getCheckType() string {
	return DOWNLOAD_CHECK_PROPMT
}

func (d *DownloadCheck) requestCheck() BosProbeErrorCode {
	if d.fileUrl != "" && d.bucketName != "" {
		return LOCAL_ARGS_BOTH_URL_BUCKET_EXIST
	}
	if d.fileUrl != "" && d.objectName != "" {
		return LOCAL_ARGS_BOTH_URL_OBJECT_EXIST
	}
	if d.fileUrl != "" && d.endpoint != "" {
		return LOCAL_ARGS_BOTH_URL_ENDPOINT_EXIST
	}
	if d.fileUrl == "" && d.bucketName == "" {
		return LOCAL_ARGS_NO_BUCKET_OR_URL
	}

	// check whether url is rightful
	if d.fileUrl != "" {
		_, err := url.Parse(d.fileUrl)
		if err != nil {
			return LOCAL_PROBE_URL_IS_INVALID
		}
	}
	return CODE_SUCCESS
}

func (d *DownloadCheck) getEndpoint(cli *bos.Client) (string, BosProbeErrorCode, error) {
	if d.endpoint != "" {
		return d.endpoint, CODE_SUCCESS, nil
	}

	if d.fileUrl != "" {
		retURL, urlErr := util.GetHostFromUrl(d.fileUrl)
		if urlErr != nil {
			return "", LOCAL_PROBE_URL_IS_INVALID, urlErr
		} else {
			return retURL, CODE_SUCCESS, nil
		}
	}

	if d.bucketName == "" {
		return "", LOCAL_ARGS_NO_BUCKET_OR_URL, fmt.Errorf("bucketName is empty")
	}

	retEndPiont, err := boscmd.GetEndpointOfBucket(cli, d.bucketName)
	if err == nil {
		return retEndPiont, CODE_SUCCESS, nil
	} else {
		return "", LOCAL_GET_ENDPOINT_BUCKET_FAILED, err
	}
}

func (d *DownloadCheck) downloadFile(sdk goSdk, fileUrl, bucketName, objectName,
	localPath string) (*serviceResp, string, error) {

	var (
		resp *serviceResp
		err  error
	)

	if fileUrl != "" {
		if localPath == "" {
			u, err := url.Parse(fileUrl)
			if err != nil {
				localPath = generateRandomFileName()
			} else {
				localPath = path.Base(u.Path)
			}
		}
		resp, err = sdk.getObjectFromUrl(fileUrl, localPath)
		return resp, localPath, err
	}

	// if user don't specify objectName, bosprobe will download the first object of bucket
	// TODO: the first object maybe too large or too small
	if objectName == "" {
		resp, objectName, err = sdk.getOneObjectFromBucket(bucketName)
		if err != nil {
			return resp, "", err
		}
	}
	tempName := path.Base(objectName)
	if localPath == "" {
		localPath = tempName
	} else {
		if util.DoesDirExist(localPath) {
			if strings.HasSuffix(localPath, util.OsPathSeparator) {
				localPath += tempName
			} else {
				localPath += util.OsPathSeparator + tempName
			}
		} else {
			if strings.HasSuffix(localPath, util.OsPathSeparator) {
				err := util.TryMkdir(localPath)
				if err != nil {
					return nil, "", err
				}
				localPath += tempName
			} else {
				dirToMake := filepath.Dir(localPath)
				err := util.TryMkdir(dirToMake)
				if err != nil {
					return nil, "", err
				}
			}
		}
	}

	resp, err = sdk.getObject(bucketName, objectName, localPath)
	return resp, localPath, err
}

func (d *DownloadCheck) checkImpl(sdk goSdk) (*serviceResp, *objectDetail, *resultStatic) {
	var (
		checkResult = &resultStatic{}
	)
	//start download test
	start := time.Now().UnixNano()
	downloadResp, localPath, downloadErr := d.downloadFile(sdk, d.fileUrl, d.bucketName,
		d.objectName, d.localPath)
	end := time.Now().UnixNano()
	useTimeMillisecond := (end - start) / 1e6

	if downloadErr != nil {
		if severErr, ok := downloadErr.(*bce.BceServiceError); ok {
			checkResult.code = BosProbeErrorCode(severErr.Code)
		} else {
			checkResult.code = boscmd.LOCAL_BCECLIENTERROR
		}
		checkResult.err = downloadErr
		return downloadResp, nil, checkResult
	} else {
		// get deatil info of the downloaded file
		objectInfo := &objectDetail{
			objectName:         localPath,
			useTimeMillisecond: useTimeMillisecond,
		}
		objectSize, getSizeErr := util.GetSizeOfFile(localPath)
		if getSizeErr != nil {
			objectSize = -1
		}
		objectInfo.objectSize = objectSize

		checkResult.code = CODE_SUCCESS
		return downloadResp, objectInfo, checkResult
	}
}
