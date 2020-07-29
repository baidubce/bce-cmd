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
	"testing"
)

import (
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
	"utils/util"
)

type shouldRetryType struct {
	err error
	ret bool
}

func TestShouldRetry(t *testing.T) {
	testCases := []shouldRetryType{
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_INVALID_ACCESS_KEY_ID,
			},
			ret: true,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_ACCESS_DENIED,
			},
			ret: true,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_NO_SUCH_BUCKET,
			},
			ret: true,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_NO_SUCH_KEY,
			},
			ret: true,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_INVALID_OBJECT_NAME,
			},
			ret: true,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: boscmd.CODE_INVALID_ARGUMENT,
			},
			ret: false,
		},
		shouldRetryType{
			err: &bce.BceServiceError{
				Code: "xxx",
			},
			ret: true,
		},
		shouldRetryType{
			err: fmt.Errorf("xxx"),
			ret: false,
		},
	}
	for i, tCase := range testCases {
		ret := shouldRetry(tCase.err)
		util.ExpectEqual("retry.go shouldRetry I", i+1, t.Errorf, tCase.ret, ret)
	}
}
