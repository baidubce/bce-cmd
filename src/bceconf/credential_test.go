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

package bceconf

import (
	// 	"fmt"
	// 	"os"
	"testing"
)

import (
	"utils/util"
)

var (
	defaultCredentialProvider, _ = NewDefaultCredentialProvider()

	fileCredentialProvider1 = &FileCredentialProvider{
		cfg: &CredentialCfg{
			Defaults: CredentialDefaultsCfg{
				Ak: "123",
				Sk: "456",
			},
		},
	}
	fileCredentialProvider2 = &FileCredentialProvider{
		cfg: &CredentialCfg{},
	}
)

type newFileCredentialProviderType struct {
	path  string
	isSuc bool
}

func TestNewFileCredentialProvider(t *testing.T) {
	testCases := []newFileCredentialProviderType{
		newFileCredentialProviderType{
			path:  "test_cred.cfg",
			isSuc: true,
		},
		newFileCredentialProviderType{
			path:  "server_test.go",
			isSuc: false,
		},
		newFileCredentialProviderType{
			isSuc: true,
		},
	}
	for i, tCase := range testCases {
		ret, err := NewFileCredentialProvider(tCase.path)
		util.ExpectEqual("credential.go NewFileCredentialProvider I", i+1, t.Errorf, tCase.isSuc,
			err == nil)
		if tCase.isSuc {
			util.ExpectEqual("credential.go NewFileCredentialProvider I", i+1, t.Errorf, true,
				ret != nil)
		}
	}
}

type credentialTestType struct {
	provider *FileCredentialProvider
	ret      string
	isSuc    bool
}

func TestGetAccessKey(t *testing.T) {
	testCases := []credentialTestType{
		credentialTestType{
			provider: fileCredentialProvider1,
			ret:      "123",
			isSuc:    true,
		},
		credentialTestType{
			provider: fileCredentialProvider2,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetAccessKey()
		util.ExpectEqual("credential.go GetAccessKey I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("credential.go GetAccessKey II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

func TestGetSecretKey(t *testing.T) {
	testCases := []credentialTestType{
		credentialTestType{
			provider: fileCredentialProvider1,
			ret:      "456",
			isSuc:    true,
		},
		credentialTestType{
			provider: fileCredentialProvider2,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetSecretKey()
		util.ExpectEqual("credential.go GetSecretKey I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("credential.go GetSecretKey II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type setCredentialType struct {
	provider *FileCredentialProvider
	val      string
	dirty    bool
}

func TestSetAccessKey(t *testing.T) {
	testCases := []setCredentialType{
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "123",
			dirty:    false,
		},
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "",
			dirty:    true,
		},
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "123",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetAccessKey(tCase.val)
		util.ExpectEqual("credential.go SetAccessKey I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("credential.go SetAccessKey II", i+1, t.Errorf, tCase.val,
			tCase.provider.cfg.Defaults.Ak)
	}
}

func TestSetSecretKey(t *testing.T) {
	testCases := []setCredentialType{
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "456",
			dirty:    false,
		},
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "",
			dirty:    true,
		},
		setCredentialType{
			provider: fileCredentialProvider1,
			val:      "456",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetSecretKey(tCase.val)
		util.ExpectEqual("credential.go SetSecretKey I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("credential.go SetSecretKey II", i+1, t.Errorf, tCase.val,
			tCase.provider.cfg.Defaults.Sk)
	}
}

type credentialSaveType struct {
	provider *FileCredentialProvider
	path     string
	setDirty bool
	isErr    bool
	dirty    bool
}

func TestCredentialSave(t *testing.T) {
	fileCredentialProvider1.dirty = false
	fileCredentialProvider2.dirty = false
	testCases := []credentialSaveType{
		credentialSaveType{
			provider: fileCredentialProvider1,
			path:     "",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		credentialSaveType{
			provider: fileCredentialProvider1,
			path:     "/root/cfg",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		credentialSaveType{
			provider: fileCredentialProvider1,
			path:     "./test.cfg",
			setDirty: true,
			isErr:    false,
			dirty:    false,
		},
		credentialSaveType{
			provider: fileCredentialProvider1,
			path:     "./test.cfg",
			setDirty: false,
			isErr:    false,
			dirty:    false,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = tCase.setDirty
		tCase.provider.configFilePath = tCase.path
		err := tCase.provider.save()
		util.ExpectEqual("credential.go save I", i+1, t.Errorf, tCase.isErr, err != nil)
		if tCase.isErr == false && err != nil {
			t.Logf("error: %s", err)
		}
		util.ExpectEqual("credential.go save II", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
	}
}

func TestDGetAccessKey(t *testing.T) {
	ret, ok := defaultCredentialProvider.GetAccessKey()
	util.ExpectEqual("credential.go de GetAccessKey I", 1, t.Errorf, true, ok)
	util.ExpectEqual("credential.go de GetAccessKey II", 1, t.Errorf, DEFAULT_AK, ret)
}

func TestDGetSecretKey(t *testing.T) {
	ret, ok := defaultCredentialProvider.GetSecretKey()
	util.ExpectEqual("credential.go de GetSecretKey I", 1, t.Errorf, true, ok)
	util.ExpectEqual("credential.go de GetSecretKey II", 1, t.Errorf, DEFAULT_SK, ret)
}

var (
	chainCredentialProvider1 = NewChainCredentialProvider([]CredentialProviderInterface{
		fileCredentialProvider1, defaultCredentialProvider})
	chainCredentialProvider2 = NewChainCredentialProvider([]CredentialProviderInterface{
		fileCredentialProvider2, defaultCredentialProvider})
)

type chainCredentialType struct {
	provider *ChainCredentialProvider
	ret      string
}

func TestCGetAccessKey(t *testing.T) {
	testCases := []chainCredentialType{
		chainCredentialType{
			provider: chainCredentialProvider1,
			ret:      "123",
		},
		chainCredentialType{
			provider: chainCredentialProvider2,
			ret:      DEFAULT_AK,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetAccessKey()
		util.ExpectEqual("server.go ch GetAccessKey I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetAccessKey II", i+1, t.Errorf, tCase.ret, ret)
	}
}

func TestCGetSecretKey(t *testing.T) {
	testCases := []chainCredentialType{
		chainCredentialType{
			provider: chainCredentialProvider1,
			ret:      "456",
		},
		chainCredentialType{
			provider: chainCredentialProvider2,
			ret:      DEFAULT_SK,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetSecretKey()
		util.ExpectEqual("server.go ch GetSecretKey I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetSecretKey II", i+1, t.Errorf, tCase.ret, ret)
	}
}
