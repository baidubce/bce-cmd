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
	"bcecmd/boscmd"
	"github.com/baidubce/bce-sdk-go/bce"
)

// Should retry when raise error?
func shouldRetry(err error) bool {

	// TODO need to consider which error code should retry
	if serverErr, ok := err.(*bce.BceServiceError); ok {
		switch serverErr.Code {
		case boscmd.CODE_INVALID_ACCESS_KEY_ID:
			return true
		case boscmd.CODE_ACCESS_DENIED:
			return true
		case boscmd.CODE_NO_SUCH_BUCKET:
			return true
		case boscmd.CODE_NO_SUCH_KEY:
			return true
		case boscmd.CODE_INVALID_OBJECT_NAME:
			return true
		case boscmd.CODE_INVALID_ARGUMENT:
			return false
		default:
			return true
		}
	} else {
		return false
	}
}
