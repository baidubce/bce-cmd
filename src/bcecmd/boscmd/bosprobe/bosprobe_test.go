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
	"path/filepath"
	"testing"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"utils/util"
)

func initConfig() error {
	userPath, err := util.GetHomeDirOfUser()
	if err != nil {
		return fmt.Errorf("Get Home of user failed")
	}
	testConfPath := filepath.Join(userPath, bceconf.DEFAULT_FOLDER_IN_USER_HOME)
	bceconf.InitConfig(testConfPath)
	return nil
}

type CheckerFake struct {
	// 1 init err, 2 request check Error
	// 3 gentEndpiont Error
	// 4 checkImp ServerError
	// 5 checkImp ClientError
	// 6 net test failed
	// 7 unknow error
	kind int
}

func (c *CheckerFake) getCheckType() string {
	return "TESTXX***DD|B"
}

func (c *CheckerFake) requestInit(req ProbeRequest) BosProbeErrorCode {
	if c.kind == 1 {
		return PROBE_INIT_REQUEST_FAILED
	}
	return CODE_SUCCESS
}

func (c *CheckerFake) requestCheck() BosProbeErrorCode {
	if c.kind == 2 {
		return LOCAL_ARGS_BOTH_URL_BUCKET_EXIST
	}
	return CODE_SUCCESS
}

func (c *CheckerFake) getEndpoint(*bos.Client) (string, BosProbeErrorCode, error) {
	if c.kind == 3 {
		return "", boscmd.CODE_NO_SUCH_KEY, &bce.BceServiceError{
			Code:    boscmd.CODE_NO_SUCH_BUCKET,
			Message: "no such bucket",
		}
	}
	if c.kind == 6 {
		return "32423xx.bcebos.com", CODE_SUCCESS, nil
	}
	return "bj.bcebos.com", CODE_SUCCESS, nil
}

func (c *CheckerFake) checkImpl(goSdk) (*serviceResp, *objectDetail, *resultStatic) {
	if c.kind == 4 {
		return &serviceResp{
				status:    404,
				requestId: "213232343243",
				debugId:   "1234567",
			}, nil, &resultStatic{
				code:      boscmd.CODE_NO_SUCH_BUCKET,
				netStatus: NET_IS_OK,
				err: &bce.BceServiceError{
					Code:    boscmd.CODE_NO_SUCH_BUCKET,
					Message: "no such bucket",
				},
			}
	} else if c.kind == 5 {
		return nil, nil, &resultStatic{
			code:      boscmd.LOCAL_BCECLIENTERROR,
			netStatus: NET_IS_OK,
			err:       fmt.Errorf("local error"),
		}
	} else if c.kind == 7 {
		return nil, nil, &resultStatic{
			code:      "TestUnkonwError",
			netStatus: NET_IS_OK,
			err:       fmt.Errorf("local error"),
		}

	}

	return &serviceResp{
			status:    200,
			requestId: "213232343243",
			debugId:   "1234567",
		}, &objectDetail{
			objectName:         "testObject",
			objectSize:         1000,
			useTimeMillisecond: 10,
		}, &resultStatic{
			code: CODE_SUCCESS,
		}
}

type handlerType struct {
	check Checker
	req   ProbeRequest
}

// Test requestInit
func TestHandler(t *testing.T) {
	handlerCases := []handlerType{
		handlerType{
			check: &CheckerFake{kind: 1},
		},
		handlerType{
			check: &CheckerFake{kind: 2},
		},
		handlerType{
			check: &CheckerFake{kind: 3},
		},
		handlerType{
			check: &CheckerFake{kind: 4},
		},
		handlerType{
			check: &CheckerFake{kind: 5},
		},
		handlerType{
			check: &CheckerFake{kind: 6},
		},
		handlerType{
			check: &CheckerFake{kind: 7},
		},
	}
	req := &DownloadCheckRes{
		BucketName: "liupeng-bj",
		Ak:         "131",
		Sk:         "131",
	}
	for _, hCase := range handlerCases {
		Handler(hCase.check, req)
	}
}
