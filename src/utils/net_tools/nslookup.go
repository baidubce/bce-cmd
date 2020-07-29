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

package net_tools

import (
	"os/exec"
	"unicode/utf8"
)

import (
	"github.com/axgle/mahonia"
)

func Nslookup(address *string) (string, error) {
	lookupOut, err := exec.Command("nslookup", *address).Output()

	ret := string(lookupOut)
	// TODO we need write a nslookup instead using system command
	if !utf8.ValidString(ret) {
		ret = mahonia.NewDecoder("gbk").ConvertString(ret)
	}
	return ret, err
}
