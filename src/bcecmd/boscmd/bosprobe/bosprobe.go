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
	"math/rand"
	"strconv"
	"time"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"utils/util"
)

var (
	DEFAULT_AUTHORITATIVE_ADRESS = "www.baidu.com"
	DEFAULT_ENDPOINT             = "bj.bcebos.com"
	PROBE_AGENT                  = "bce-bos-probe"
)

var (
	pLog *probeLog
)

// the interface of upload and download checker
type Checker interface {
	getCheckType() string
	requestInit(req ProbeRequest) BosProbeErrorCode
	requestCheck() BosProbeErrorCode
	getEndpoint(*bos.Client) (string, BosProbeErrorCode, error)
	checkImpl(goSdk) (*serviceResp, *objectDetail, *resultStatic)
}

// the interface of request
type ProbeRequest interface {
	GetAk() string
	GetSk() string
	GetEndPoint() string
}

type UploadCheckRes struct {
	Ak         string
	Sk         string
	BucketName string
	ObjectName string
	LocalPath  string
	Endpoint   string
}

func (u *UploadCheckRes) GetAk() string {
	return u.Ak
}

func (u *UploadCheckRes) GetSk() string {
	return u.Sk
}

func (u *UploadCheckRes) GetEndPoint() string {
	return u.Endpoint
}

type DownloadCheckRes struct {
	Ak         string
	Sk         string
	FileUrl    string
	BucketName string
	ObjectName string
	LocalPath  string
	Endpoint   string
}

func (d *DownloadCheckRes) GetAk() string {
	return d.Ak
}

func (d *DownloadCheckRes) GetSk() string {
	return d.Sk
}

func (d *DownloadCheckRes) GetEndPoint() string {
	return d.Endpoint
}

// record the detail info of uploaded or downloaded file
type objectDetail struct {
	objectName         string
	objectSize         int64
	useTimeMillisecond int64
}

// the results of upload or download
type resultStatic struct {
	code      BosProbeErrorCode
	netStatus NetStatusCode
	err       error
}

func generateRandomFileName() string {
	return "probe" + time.Now().Format("20060102150405") + strconv.Itoa(rand.Intn(1000)) + ".temp"
}

func probeinit() {
	rand.Seed(time.Now().UnixNano())
	pLog = &probeLog{}
	pLog.createLogFile()
}

// the entry of upload test
func Handler(c Checker, req ProbeRequest) {
	var (
		checkCode   BosProbeErrorCode = PROBE_NOT_CHECK
		netStatus   NetStatusCode     = NET_NOT_CHECK
		bosClient   *bos.Client
		uploadResp  *serviceResp
		objectInfo  *objectDetail
		checkResult *resultStatic
		sdk         goSdk
		ak          string
		sk          string
		endpoint    string
		checkType   string
		err         error
	)
	stopChan := make(chan bool)

	probeinit()

	// get cmd detail and operator system info
	getBasicInfo()

	checkType = c.getCheckType()

	// init request
	checkCode = c.requestInit(req)
	if checkCode != CODE_SUCCESS {
		goto SUGGESTION
	}

	//request check
	checkCode = c.requestCheck()
	if checkCode != CODE_SUCCESS {
		goto SUGGESTION
	}

	ak = req.GetAk()
	sk = req.GetSk()
	endpoint = req.GetEndPoint()

	bosClient, err = NewClient(ak, sk, endpoint)
	if err != nil {
		checkCode = boscmd.LOCAL_INIT_BOSCLIENT_FAILED
		goto SUGGESTION
	}

	// if endpoint is empty, get it from url or bucketname
	if endpoint == "" {
		endpoint, checkCode, err = c.getEndpoint(bosClient)
		if err != nil {
			goto SUGGESTION
		} else {
			bosClient.Config.Endpoint = endpoint
		}
	}

	sdk = NewGoSdk(bosClient)

	//get network status
	checkCode = PROBE_NOT_CHECK
	go util.PrintWaiting("网络检测中....", stopChan)
	netStatus = NetStatusReport(endpoint, DEFAULT_AUTHORITATIVE_ADRESS)
	stopChan <- true
	if netStatus != NET_IS_OK {
		fmt.Printf("[网络不通]\n")
		goto SUGGESTION
	} else {
		fmt.Printf("[网络良好]\n")
	}

	// UploadCheckImpl will test net status and test upload file
	go util.PrintWaiting(checkType+"检测中....", stopChan)
	uploadResp, objectInfo, checkResult = c.checkImpl(sdk)
	stopChan <- true
	checkCode = checkResult.code
	if checkCode == CODE_SUCCESS {
		fmt.Printf("[%s成功]\n", checkType)
		if objectInfo != nil {
			printCheckDetail(checkType, objectInfo)
		}
	} else {
		fmt.Printf("[%s失败]\n", checkType)
	}
	err = checkResult.err

SUGGESTION:
	// print check result (success or fail)
	printCheckResult(checkType, netStatus, checkCode)
	// print net status
	if netStatus != NET_NOT_CHECK {
		PrintNetStatus(netStatus)
	}
	// print the error detail
	if checkCode != CODE_SUCCESS {
		PrintFailedDetail(checkCode, err, true)
	}
	// give suggestion
	giveSuggestion(netStatus, checkCode, err)
	// save debug info
	if uploadResp != nil {
		PrintDebugInfo(uploadResp)
	}
	PrintLogInfo()
}
