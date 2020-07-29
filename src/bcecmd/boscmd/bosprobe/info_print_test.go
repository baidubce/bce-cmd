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
	"testing"
)

import (
	"bcecmd/boscmd"
	"utils/util"
)

type disturbAkSkType struct {
	cmd   string
	exCmd string
}

// Test disturbAkSk
func TestDisturbAkSk(t *testing.T) {
	disturbAkSkCases := []disturbAkSkType{
		disturbAkSkType{
			cmd:   "cmd -a1234",
			exCmd: "cmd -a***",
		},
		disturbAkSkType{
			cmd:   "cmd -s1234",
			exCmd: "cmd -s***",
		},
		disturbAkSkType{
			cmd:   "cmd -s1234 -a 123",
			exCmd: "cmd -s*** -a ***",
		},
	}
	for i, dCase := range disturbAkSkCases {
		ret := disturbAkSk(dCase.cmd)
		util.ExpectEqual("info_print disturbAkSk ", i, t.Errorf, dCase.exCmd, ret)
	}
}

type netStatusReportType struct {
	endpoint    string
	authAddress string
	exCode      NetStatusCode
}

// Test disturbAkSk
func TestNetStatusReport(t *testing.T) {
	netStatusReportCases := []netStatusReportType{
		netStatusReportType{
			endpoint:    "bj.bcebos.com",
			authAddress: "baidu.com",
			exCode:      NET_IS_OK,
		},
		netStatusReportType{
			endpoint:    "bj.bcebos.com",
			authAddress: "www.qq.com",
			exCode:      NET_IS_OK,
		},
		netStatusReportType{
			endpoint:    "bjxxx.bcebos.com",
			authAddress: "www.qq.com",
			exCode:      CLIENT_NET_ERROR,
		},
		netStatusReportType{
			endpoint:    "bjxxx.bcebos.com",
			authAddress: "bj.bcebos.com",
			exCode:      ENDPOINT_CONNECT_ERROR,
		},
	}
	for i, nCase := range netStatusReportCases {
		ret := NetStatusReport(nCase.endpoint, nCase.authAddress)
		util.ExpectEqual("info_print NetStatusReport ", i, t.Errorf, nCase.exCode, ret)
	}
}

func TestPrintRequestCheckInfo(t *testing.T) {
	printRequestCheckInfo(CODE_SUCCESS)
	printRequestCheckInfo(boscmd.CODE_NO_SUCH_KEY)
}

func TestPrintNetStatus(t *testing.T) {
	PrintNetStatus(NET_NOT_CHECK)
	PrintNetStatus(NET_IS_OK)
	PrintNetStatus(ENDPOINT_CONNECT_ERROR)
	PrintNetStatus(CLIENT_NET_ERROR)
}
