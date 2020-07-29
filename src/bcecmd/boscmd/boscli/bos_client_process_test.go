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
	// 	"fmt"
	"os"
	// 	"runtime"
	"strings"
	"testing"
	// 	"path/filepath"
)

import (
	"bceconf"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"utils/util"
)

func init() {
	if err := initConfig(); err != nil {
		os.Exit(1)
	}
}

func TestGetUserAgent(t *testing.T) {
	ret := getUserAgent()
	util.ExpectEqual("process.go getUserAgent I", 1, t.Errorf, false, ret == "")
}

type setEndpointProtocolType struct {
	input    string
	output   string
	useHttps bool
	isSuc    bool
}

func TestSetEndpointProtocol(t *testing.T) {
	testCases := []setEndpointProtocolType{
		setEndpointProtocolType{
			input:    "bj.bcebos.com",
			useHttps: false,
			output:   "http://bj.bcebos.com",
			isSuc:    true,
		},
		setEndpointProtocolType{
			input:    "bj.bcebos.com",
			useHttps: true,
			output:   "https://bj.bcebos.com",
			isSuc:    true,
		},
		setEndpointProtocolType{
			input:    "http://bj.bcebos.com",
			useHttps: true,
			output:   "https://bj.bcebos.com",
			isSuc:    true,
		},
		setEndpointProtocolType{
			input:    "https://bj.bcebos.com",
			useHttps: true,
			output:   "https://bj.bcebos.com",
			isSuc:    true,
		},
		setEndpointProtocolType{
			input:    "http://bj.bcebos.com",
			useHttps: false,
			output:   "http://bj.bcebos.com",
			isSuc:    true,
		},

		setEndpointProtocolType{
			input:    "https://bj.bcebos.com",
			useHttps: false,
			output:   "http://bj.bcebos.com",
			isSuc:    true,
		},
		setEndpointProtocolType{
			input:    "",
			useHttps: true,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, err := setEndpointProtocol(tCase.input, tCase.useHttps)
		util.ExpectEqual("tools.go setEndpointProtocol I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("tools.go setEndpointProtocol II", i+1, t.Errorf, tCase.output, ret)
		}
	}
}

type newBosClientType struct {
	input string
	isSuc bool
}

func TestNewBosClient(t *testing.T) {
	testCases := []newBosClientType{
		newBosClientType{
			input: "bj.bcebos.com",
			isSuc: true,
		},
		newBosClientType{
			input: "",
			isSuc: false,
		},
		newBosClientType{
			input: "http://bj",
			isSuc: true,
		},
	}
	for i, tCase := range testCases {
		_, err := newBosClient("", "", tCase.input, false, 10, 10)
		util.ExpectEqual("tools.go newBosClient I", i, t.Errorf, tCase.isSuc, err == nil)
		if err != nil {
			t.Logf("err; %s", err)
		}
	}
}

type buildBosClientType struct {
	ak                   string
	sk                   string
	endpoint             string
	providerAk           string
	providerSk           string
	providerEndpoint     string
	providerUseHttps     string
	providerSetHttp      bool
	multiUploadThreadNum string
	multiUploadPartSize  string
	finAk                string
	finSk                string
	outEndpoint          string
	isSuc                bool
}

func TestBuildBosClient(t *testing.T) {
	testCases := []buildBosClientType{
		buildBosClientType{
			ak:                   "",
			sk:                   "",
			endpoint:             "",
			providerAk:           "123",
			providerSk:           "456",
			providerEndpoint:     "xxx.com",
			providerUseHttps:     "yes",
			providerSetHttp:      true,
			multiUploadThreadNum: "10",
			multiUploadPartSize:  "13",
			finAk:                "123",
			finSk:                "456",
			outEndpoint:          "https://xxx.com",
			isSuc:                true,
		},
		buildBosClientType{
			ak:                   "",
			sk:                   "",
			endpoint:             "",
			providerAk:           "",
			providerSk:           "",
			providerEndpoint:     "",
			providerUseHttps:     "",
			providerSetHttp:      false,
			multiUploadThreadNum: "10",
			multiUploadPartSize:  "13",
			finAk:                "",
			finSk:                "",
			outEndpoint:          "",
			isSuc:                false,
		},
		buildBosClientType{
			ak:                   "",
			sk:                   "",
			endpoint:             "",
			providerAk:           "123",
			providerSk:           "",
			providerEndpoint:     "",
			providerUseHttps:     "",
			multiUploadThreadNum: "10",
			multiUploadPartSize:  "13",
			providerSetHttp:      false,
			finAk:                "",
			finSk:                "",
			outEndpoint:          "",
			isSuc:                false,
		},
		buildBosClientType{
			ak:                   "",
			sk:                   "",
			endpoint:             "",
			providerAk:           "123",
			providerSk:           "456",
			providerEndpoint:     "",
			providerUseHttps:     "",
			providerSetHttp:      false,
			multiUploadThreadNum: "10",
			multiUploadPartSize:  "13",
			finAk:                "",
			finSk:                "",
			outEndpoint:          "",
			isSuc:                false,
		},
		buildBosClientType{
			ak:                  "",
			sk:                  "",
			endpoint:            "xxx.com",
			providerAk:          "123",
			providerSk:          "456",
			providerEndpoint:    "",
			providerUseHttps:    "",
			providerSetHttp:     true,
			multiUploadPartSize: "",
			finAk:               "",
			finSk:               "",
			outEndpoint:         "",
			isSuc:               false,
		},
	}

	creProvider, err := bceconf.NewFileCredentialProvider("./config/credentials")
	if err != nil {
		t.Errorf("init NewFileCredentialProvider failed")
		return
	}
	serverProvider, err := bceconf.NewFileServerConfigProvider("./config/config")
	if err != nil {
		t.Errorf("init NewFileServerConfigProvider failed")
		return
	}
	for i, tCase := range testCases {
		creProvider.SetAccessKey(tCase.providerAk)
		creProvider.SetSecretKey(tCase.providerSk)
		serverProvider.SetDomain(tCase.providerEndpoint)
		serverProvider.SetUseHttpsProtocol(tCase.providerUseHttps)
		serverProvider.SetMultiUploadPartSize(tCase.multiUploadPartSize)
		serverProvider.SetMultiUploadThreadNum(tCase.multiUploadThreadNum)

		ret, err := buildBosClient(tCase.ak, tCase.sk, tCase.endpoint, creProvider, serverProvider)
		util.ExpectEqual("process.go buildBosClient I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if !tCase.isSuc || err != nil {
			t.Logf("err; %s", err)
			continue
		}
		util.ExpectEqual("process.go buildBosClient II", i+1, t.Errorf, tCase.finAk,
			ret.Config.Credentials.AccessKeyId)
		util.ExpectEqual("process.go buildBosClient III", i+1, t.Errorf, tCase.finSk,
			ret.Config.Credentials.SecretAccessKey)
		util.ExpectEqual("process.go buildBosClient IV", i+1, t.Errorf, tCase.outEndpoint,
			ret.Config.Endpoint)
	}
}

type bosClientInitType struct {
	useAuto string
	isWrap  bool
	isSuc   bool
}

func TestBosClientInit(t *testing.T) {
	testCases := []bosClientInitType{
		bosClientInitType{
			useAuto: "yes",
			isWrap:  true,
			isSuc:   true,
		},
		bosClientInitType{
			useAuto: "y",
			isWrap:  true,
			isSuc:   true,
		},
		bosClientInitType{
			useAuto: "n",
			isWrap:  false,
			isSuc:   true,
		},
		bosClientInitType{
			useAuto: "no",
			isWrap:  false,
			isSuc:   true,
		},
	}

	temp := bceconf.ServerConfigProvider
	// generate server config provider
	serverConfigFileProvider, err := bceconf.NewFileServerConfigProvider("./config/config")
	if err != nil {
		t.Errorf("init NewFileServerConfigProvider failed")
		return
	}
	defaServerConfigProvider, err := bceconf.NewDefaultServerConfigProvider()
	if err != nil {
		t.Errorf("init NewDefaultServerConfigProvider failed")
		return
	}
	bceconf.ServerConfigProvider = bceconf.NewChainServerConfigProvider(
		[]bceconf.ServerConfigProviderInterface{serverConfigFileProvider,
			defaServerConfigProvider})

	for i, tCase := range testCases {
		ok := serverConfigFileProvider.SetUseAutoSwitchDomain(tCase.useAuto)
		if !ok {
			t.Errorf("set SetUseAutoSwitchDomain failed %s", tCase.useAuto)
			continue
		}
		ret, err := bosClientInit("", "", "")
		if !tCase.isSuc {
			util.ExpectEqual("process.go buildBosClient I", i+1, t.Errorf, tCase.isSuc, err == nil)
		} else {
			_, ok := ret.(*bosClientWrapper)
			util.ExpectEqual("process.go buildBosClient II", i+1, t.Errorf, tCase.isWrap, ok)
		}
	}
	bceconf.ServerConfigProvider = temp
}

type initBosClientForBucketType struct {
	ak       string
	sk       string
	bucket   string
	endpoint string
	isSuc    bool
}

func TestInitBosClientForBucket(t *testing.T) {
	testCases := []initBosClientForBucketType{
		//1
		initBosClientForBucketType{
			bucket:   "liupeng-bj",
			endpoint: "bj.bcebos.com",
			isSuc:    true,
		},
		//2
		initBosClientForBucketType{
			bucket:   "liupeng-su",
			endpoint: "su.bcebos.com",
			isSuc:    true,
		},
		//3
		initBosClientForBucketType{
			bucket:   "liupeng-gz",
			endpoint: "gz.bcebos.com",
			isSuc:    true,
		},
		//4
		initBosClientForBucketType{
			bucket:   "liupeng-hk02",
			endpoint: "hk-2.bcebos.com",
			isSuc:    true,
		},
		//5
		// get endpoint from cache
		initBosClientForBucketType{
			ak:     "123",
			sk:     "456",
			bucket: "liupeng-gz",
			isSuc:  true,
		},
		//6
		initBosClientForBucketType{
			ak:     "123",
			sk:     "456",
			bucket: "liupeng-xx",
			isSuc:  false,
		},
	}

	temp := bceconf.ServerConfigProvider
	// generate server config provider
	serverConfigFileProvider, err := bceconf.NewFileServerConfigProvider("./config/config")
	if err != nil {
		t.Errorf("init NewFileServerConfigProvider failed")
		return
	}
	defaServerConfigProvider, err := bceconf.NewDefaultServerConfigProvider()
	if err != nil {
		t.Errorf("init NewDefaultServerConfigProvider failed")
		return
	}
	bceconf.ServerConfigProvider = bceconf.NewChainServerConfigProvider(
		[]bceconf.ServerConfigProviderInterface{serverConfigFileProvider,
			defaServerConfigProvider})

	serverConfigFileProvider.SetUseAutoSwitchDomain("no")
	endpoint, _ := bceconf.ServerConfigProvider.GetDomain()
	useHttps, _ := bceconf.ServerConfigProvider.GetUseHttpsProtocol()
	endpoint, err = setEndpointProtocol(endpoint, useHttps)
	if err != nil {
		t.Errorf("TestInitBosClientForBucket: get originl endpoint  failed")
	}

	for i, tCase := range testCases {
		ret, err := initBosClientForBucket(tCase.ak, tCase.sk, tCase.bucket)
		util.ExpectEqual("process.go initBosClientForBucket I", i+1, t.Errorf, true, err == nil)
		bosClient, ok := ret.(*bos.Client)
		util.ExpectEqual("process.go initBosClientForBucket III", i+1, t.Errorf, true, ok)
		retEndpoint := bosClient.Config.Endpoint
		util.ExpectEqual("process.go initBosClientForBucket II", i+1, t.Errorf, true,
			strings.HasSuffix(retEndpoint, endpoint))
		t.Logf("endpoint: %s", retEndpoint)
	}

	serverConfigFileProvider.SetUseAutoSwitchDomain("yes")
	for i, tCase := range testCases {
		ret, err := initBosClientForBucket(tCase.ak, tCase.sk, tCase.bucket)
		if !tCase.isSuc {
			util.ExpectEqual("process.go initBosClientForBucket III", i+1, t.Errorf, tCase.isSuc,
				err == nil)
		} else {
			bosClient, _ := ret.(*bosClientWrapper)
			retEndpoint := bosClient.bosClient.Config.Endpoint
			util.ExpectEqual("process.go initBosClientForBucket IV", i+1, t.Errorf, true,
				strings.HasSuffix(retEndpoint, tCase.endpoint))
			t.Logf("endpoint: %s", retEndpoint)
		}
	}
	bceconf.ServerConfigProvider = temp
}

type modifiyBosClientEndpointByEndpointType struct {
	useHttps    string
	endpoint    string
	outEndpoint string
	isSuc       bool
}

func TestModifiyBosClientEndpointByEndpoint(t *testing.T) {
	testCases := []modifiyBosClientEndpointByEndpointType{
		modifiyBosClientEndpointByEndpointType{
			useHttps:    "yes",
			endpoint:    "bj.bcebos.com",
			outEndpoint: "https://bj.bcebos.com",
			isSuc:       true,
		},
		modifiyBosClientEndpointByEndpointType{
			useHttps:    "no",
			endpoint:    "su.bcebos.com",
			outEndpoint: "http://su.bcebos.com",
			isSuc:       true,
		},
		modifiyBosClientEndpointByEndpointType{
			useHttps:    "xx",
			endpoint:    "xx.bcebos.com",
			outEndpoint: "http://xx.bcebos.com",
			isSuc:       true,
		},
		modifiyBosClientEndpointByEndpointType{
			endpoint: "",
			isSuc:    false,
		},
	}

	temp := bceconf.ServerConfigProvider
	// generate server config provider
	serverConfigFileProvider, err := bceconf.NewFileServerConfigProvider("./config/config")
	if err != nil {
		t.Errorf("init NewFileServerConfigProvider failed")
		return
	}
	defaServerConfigProvider, err := bceconf.NewDefaultServerConfigProvider()
	if err != nil {
		t.Errorf("init NewDefaultServerConfigProvider failed")
		return
	}
	bceconf.ServerConfigProvider = bceconf.NewChainServerConfigProvider(
		[]bceconf.ServerConfigProviderInterface{serverConfigFileProvider,
			defaServerConfigProvider})

	bosClient, err := bos.NewClient("", "", "")
	if err != nil {
		t.Errorf("init bosClient failed")
		return
	}

	for i, tCase := range testCases {
		serverConfigFileProvider.SetUseHttpsProtocol(tCase.useHttps)
		err := modifiyBosClientEndpointByEndpoint(bosClient, tCase.endpoint)
		if !tCase.isSuc {
			util.ExpectEqual("process.go modifiyBosClientEndpointByEndpoint I", i+1, t.Errorf,
				tCase.isSuc, err == nil)
		} else {
			util.ExpectEqual("process.go modifiyBosClientEndpointByEndpoint II", i+1, t.Errorf,
				tCase.outEndpoint, bosClient.Config.Endpoint)
		}
	}
	bceconf.ServerConfigProvider = temp
}

type modifiyBosClientEndpointByBucketNameCacheType struct {
	ak            string
	sk            string
	bucket        string
	insertToCache bool
	endpoint      string
	isSuc         bool
}

func TestModifiyBosClientEndpointByBucketNameCache(t *testing.T) {
	testCases := []modifiyBosClientEndpointByBucketNameCacheType{
		modifiyBosClientEndpointByBucketNameCacheType{
			bucket:        "liupeng-bj",
			insertToCache: false,
			isSuc:         false,
		},
		modifiyBosClientEndpointByBucketNameCacheType{
			bucket:        "liupeng-bj",
			endpoint:      "bj.bcebos.com",
			insertToCache: true,
			isSuc:         true,
		},
		modifiyBosClientEndpointByBucketNameCacheType{
			bucket:        "liupeng-su",
			endpoint:      "su.bcebos.com",
			insertToCache: true,
			isSuc:         true,
		},
		modifiyBosClientEndpointByBucketNameCacheType{
			bucket:        "liupeng-xxx",
			insertToCache: false,
			isSuc:         false,
		},
	}

	bosClient, err := bos.NewClient("", "", "")
	if err != nil {
		t.Errorf("init bosClient failed")
		return
	}

	for i, tCase := range testCases {

		bceconf.BucketEndpointCacheProvider.Delete(tCase.bucket)
		if tCase.insertToCache {
			bceconf.BucketEndpointCacheProvider.Write(tCase.bucket, tCase.endpoint, 1800)
		}

		ok := modifiyBosClientEndpointByBucketNameCache(bosClient, tCase.bucket)
		if !tCase.isSuc {
			util.ExpectEqual("process.go modifiyBosClientEndpointByBucketNameCache I", i+1,
				t.Errorf, tCase.isSuc, ok)
		} else {
			retEndpoint := bosClient.Config.Endpoint
			util.ExpectEqual("process.go modifiyBosClientEndpointByBucketNameCache II", i+1,
				t.Errorf, true, strings.HasSuffix(retEndpoint, tCase.endpoint))
			t.Logf("endpoint: %s", retEndpoint)
		}
	}
}

type modifiyBosClientEndpointByBucketNameBosType struct {
	ak       string
	sk       string
	bucket   string
	endpoint string
	isSuc    bool
}

func TestModifiyBosClientEndpointByBucketNameBos(t *testing.T) {
	testCases := []modifiyBosClientEndpointByBucketNameBosType{
		modifiyBosClientEndpointByBucketNameBosType{
			bucket:   "liupeng-bj",
			endpoint: "bj.bcebos.com",
			isSuc:    true,
		},
		modifiyBosClientEndpointByBucketNameBosType{
			bucket:   "liupeng-gz",
			endpoint: "gz.bcebos.com",
			isSuc:    true,
		},
		modifiyBosClientEndpointByBucketNameBosType{
			bucket:   "liupeng-su",
			endpoint: "su.bcebos.com",
			isSuc:    true,
		},
		modifiyBosClientEndpointByBucketNameBosType{
			bucket: "liupeng-xxx",
			isSuc:  false,
		},
	}
	bosClient, err := buildBosClient("", "", "", bceconf.CredentialProvider,
		bceconf.ServerConfigProvider)
	if err != nil {
		t.Errorf("init bosClient failed")
		return
	}

	for i, tCase := range testCases {

		err := modifiyBosClientEndpointByBucketNameBos(bosClient, tCase.bucket)
		if !tCase.isSuc {
			util.ExpectEqual("process.go modifiyBosClientEndpointByBucketNameBos I", i+1,
				t.Errorf, tCase.isSuc, err == nil)
		} else {
			retEndpoint := bosClient.Config.Endpoint
			util.ExpectEqual("process.go modifiyBosClientEndpointByBucketNameBos II", i+1,
				t.Errorf, true, strings.HasSuffix(retEndpoint, tCase.endpoint))
			t.Logf("endpoint: %s", retEndpoint)
		}
	}
}
