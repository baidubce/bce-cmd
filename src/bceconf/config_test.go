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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

import (
	"utils/util"
)

var (
	fileServerProviderTestVal = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				Domain:                   "bj.bcebos.com",
				Region:                   "bj",
				AutoSwitchDomain:         "yes",
				BreakpointFileExpiration: "10",
				Https:                "yes",
				MultiUploadThreadNum: "20",
				MultiUploadPartSize:  "10",
				SyncProcessingNum:    "22",
			},
			Domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{
					Endpoint: "bj.bcebos.com",
				},
			},
		},
	}
	fileCredentialProviderTestVal = &FileCredentialProvider{
		cfg: &CredentialCfg{
			Defaults: CredentialDefaultsCfg{
				Ak: "123",
				Sk: "456",
			},
		},
	}
)

type initConfFolderType struct {
	path  string
	isSuc bool
}

func TestInitConfFolder(t *testing.T) {
	_, err := os.Create("initConfTest")
	if err != nil {
		t.Errorf("create file initConfTest failed")
		return
	}

	testCases := []initConfFolderType{
		initConfFolderType{
			path:  "./initConfTest",
			isSuc: false,
		},
		initConfFolderType{
			path:  "./testcfgint",
			isSuc: true,
		},
		initConfFolderType{
			path:  "/initConftest",
			isSuc: false,
		},
	}
	for i, tCase := range testCases {
		err := initConfFolder(tCase.path)
		util.ExpectEqual("config.go initConfFolder I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("config.go initConfFolder II", i+1, t.Errorf, true,
				util.DoesDirExist(tCase.path))
			if err != nil {
				t.Errorf("error: %s", err)
			}
		}
	}
}

type initConfigType struct {
	path           string
	serverProvider *FileServerConfigProvider
	credenProvider *FileCredentialProvider
	isSuc          bool
}

func TestInitConfig(t *testing.T) {
	testCases := []initConfigType{
		initConfigType{
			path:           "./initConfigTest",
			serverProvider: fileServerProviderTestVal,
			credenProvider: fileCredentialProviderTestVal,
			isSuc:          true,
		},
		initConfigType{
			path:  "/initConfigTest",
			isSuc: false,
		},
		initConfigType{
			path:  "",
			isSuc: true,
		},
	}

	for i, tCase := range testCases {
		if tCase.isSuc {
			path := tCase.path
			if tCase.path == "" {
				configDirPath, err := util.GetHomeDirOfUser()
				if err != nil {
					t.Errorf("get home dir of user failed")
					continue
				}
				path = filepath.Join(configDirPath, DEFAULT_FOLDER_IN_USER_HOME)
				serverPath := path + util.OsPathSeparator + "config"
				credenPath := path + util.OsPathSeparator + "credentials"
				tCase.serverProvider, _ = NewFileServerConfigProvider(serverPath)
				tCase.credenProvider, _ = NewFileCredentialProvider(credenPath)
			} else {
				initConfFolder(path)
				tCase.serverProvider.configFilePath = filepath.Join(tCase.path, "config")
				tCase.credenProvider.configFilePath = filepath.Join(tCase.path, "credentials")
			}
			tCase.serverProvider.dirty = true
			tCase.serverProvider.save()
			tCase.credenProvider.dirty = true
			tCase.credenProvider.save()
		}
		err := InitConfig(tCase.path)
		util.ExpectEqual("config.go InitConfig I", i+1, t.Errorf, tCase.isSuc, err == nil)
		if tCase.isSuc {
			util.ExpectEqual("config.go InitConfig II", i+1, t.Errorf,
				tCase.serverProvider.cfg.Defaults, serverConfigFileProvider.cfg.Defaults)
			if len(tCase.serverProvider.cfg.Domains) == 0 {
				util.ExpectEqual("config.go InitConfig III", i+1, t.Errorf, 0,
					len(serverConfigFileProvider.cfg.Domains))
			} else {
				util.ExpectEqual("config.go InitConfig III", i+1, t.Errorf,
					tCase.serverProvider.cfg.Domains, serverConfigFileProvider.cfg.Domains)
			}
			util.ExpectEqual("config.go InitConfig IV", i+1, t.Errorf,
				tCase.credenProvider.cfg.Defaults, credentialFileProvider.cfg.Defaults)
			t.Logf("%v", serverConfigFileProvider.cfg.Defaults)
		}
	}
}

type reloadConfActionType struct {
	path           string
	serverProvider *FileServerConfigProvider
	credenProvider *FileCredentialProvider
	isSuc          bool
}

func TestReloadConfAction(t *testing.T) {
	testCases := []reloadConfActionType{
		reloadConfActionType{
			path:           "./initConfigTest",
			serverProvider: fileServerProviderTestVal,
			credenProvider: fileCredentialProviderTestVal,
			isSuc:          true,
		},
		reloadConfActionType{
			path:  "",
			isSuc: true,
		},
	}

	for i, tCase := range testCases {
		if tCase.isSuc {
			path := tCase.path
			if tCase.path == "" {
				configDirPath, err := util.GetHomeDirOfUser()
				if err != nil {
					t.Errorf("get home dir of user failed")
					continue
				}
				path = filepath.Join(configDirPath, DEFAULT_FOLDER_IN_USER_HOME)
				serverPath := path + util.OsPathSeparator + "config"
				credenPath := path + util.OsPathSeparator + "credentials"
				tCase.serverProvider, _ = NewFileServerConfigProvider(serverPath)
				tCase.credenProvider, _ = NewFileCredentialProvider(credenPath)
			} else {
				initConfFolder(path)
				tCase.serverProvider.configFilePath = filepath.Join(tCase.path, "config")
				tCase.credenProvider.configFilePath = filepath.Join(tCase.path, "credentials")
			}
			tCase.serverProvider.dirty = true
			tCase.serverProvider.save()
			tCase.credenProvider.dirty = true
			tCase.credenProvider.save()
		}
		ReloadConfAction(tCase.path)
		util.ExpectEqual("config.go ReloadConfAction I", i+1, t.Errorf,
			tCase.serverProvider.cfg.Defaults, serverConfigFileProvider.cfg.Defaults)
		if len(tCase.serverProvider.cfg.Domains) == 0 {
			util.ExpectEqual("config.go ReloadConfAction II", i+1, t.Errorf, 0,
				len(serverConfigFileProvider.cfg.Domains))
		} else {
			util.ExpectEqual("config.go ReloadConfAction II", i+1, t.Errorf,
				tCase.serverProvider.cfg.Domains, serverConfigFileProvider.cfg.Domains)
		}
		util.ExpectEqual("config.go ReloadConfAction III", i+1, t.Errorf,
			tCase.credenProvider.cfg.Defaults, credentialFileProvider.cfg.Defaults)
		t.Logf("%v", serverConfigFileProvider.cfg.Defaults)
	}
}

func TestDestructConfFolder(t *testing.T) {
	serverConfigFileProvider = fileServerProviderTestVal
	serverConfigFileProvider.configFilePath = "./initConfigTest/config"
	if util.DoesFileExist(serverConfigFileProvider.configFilePath) {
		os.Remove(serverConfigFileProvider.configFilePath)
	}
	serverConfigFileProvider.dirty = true
	DestructConfFolder()
	util.ExpectEqual("config.go DestructConfFolder I", 1, t.Errorf,
		true, util.DoesFileExist(serverConfigFileProvider.configFilePath))
}

type configInteractiveType struct {
	path       string
	filePath   string
	ak         string
	sk         string
	region     string
	domain     string
	autoSwitch string
	breakpo    string
	httpsp     string
	mulitNum   string
	syncNum    string
	partSize   string
}

func TestConfigInteractive(t *testing.T) {
	testCases := []configInteractiveType{
		configInteractiveType{
			path:       "./test11",
			filePath:   "./test_input1",
			ak:         "123",
			sk:         "456",
			region:     "bx",
			domain:     "bx.com",
			autoSwitch: "10",
			breakpo:    "20",
			httpsp:     "yes",
			mulitNum:   "30",
			syncNum:    "40",
			partSize:   "",
		},
		configInteractiveType{
			path:       "./test12",
			filePath:   "./test_input2",
			ak:         "123",
			sk:         "456",
			region:     "none",
			domain:     "bx.com",
			autoSwitch: "10",
			breakpo:    "none",
			httpsp:     "",
			mulitNum:   "30",
			syncNum:    "",
			partSize:   "14",
		},
		configInteractiveType{
			path:       "./test13",
			filePath:   "./test_input3",
			ak:         "123",
			sk:         "456",
			region:     "none",
			domain:     "bx.com",
			autoSwitch: "ax10",
			breakpo:    "a3vd43",
			httpsp:     "xx",
			mulitNum:   "xvc30",
			syncNum:    "dsffcv",
			partSize:   "1dfds",
		},
		configInteractiveType{
			path:       "./test12",
			filePath:   "./test_input2",
			ak:         "none",
			sk:         "none",
			region:     "none",
			domain:     "none",
			autoSwitch: "none",
			breakpo:    "none",
			httpsp:     "none",
			mulitNum:   "none",
			syncNum:    "none",
			partSize:   "none",
		},
	}
	temp := os.Stdin
	for i, tCase := range testCases {
		fd, err := os.Create(tCase.filePath)
		if err != nil {
			t.Errorf("create file %s filed", tCase.filePath)
			continue
		}
		fmt.Fprintf(fd, "%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", tCase.ak, tCase.sk, tCase.region,
			tCase.domain, tCase.autoSwitch, tCase.breakpo, tCase.httpsp, tCase.mulitNum,
			tCase.syncNum, tCase.partSize)
		fd.Close()

		fd, err = os.OpenFile(tCase.filePath, os.O_RDONLY, 0755)
		if err != nil {
			t.Errorf("open file %s filed", tCase.filePath)
			continue
		}
		os.Stdin = fd
		ConfigInteractive(tCase.path)
		fd.Close()

		if tCase.region == "none" {
			util.ExpectEqual("config.go ConfigInteractive  I", i+1, t.Errorf, "",
				serverConfigFileProvider.cfg.Defaults.Region)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  I", i+1, t.Errorf, tCase.region,
				serverConfigFileProvider.cfg.Defaults.Region)
		}
		if tCase.domain == "none" {
			util.ExpectEqual("config.go ConfigInteractive  II", i+1, t.Errorf, "",
				serverConfigFileProvider.cfg.Defaults.Domain)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  II", i+1, t.Errorf, tCase.domain,
				serverConfigFileProvider.cfg.Defaults.Domain)
		}
		if tCase.ak == "none" {
			util.ExpectEqual("config.go ConfigInteractive  III", i+1, t.Errorf, "",
				credentialFileProvider.cfg.Defaults.Ak)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  III", i+1, t.Errorf, tCase.ak,
				credentialFileProvider.cfg.Defaults.Ak)
		}
		if _, ok := strconv.Atoi(tCase.breakpo); ok != nil {
			util.ExpectEqual("config.go ConfigInteractive  IV", i+1, t.Errorf, "",
				serverConfigFileProvider.cfg.Defaults.BreakpointFileExpiration)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  IV", i+1, t.Errorf, tCase.breakpo,
				serverConfigFileProvider.cfg.Defaults.BreakpointFileExpiration)
		}
		if _, ok := strconv.Atoi(tCase.mulitNum); ok != nil {
			util.ExpectEqual("config.go ConfigInteractive  V", i+1, t.Errorf, "",
				serverConfigFileProvider.cfg.Defaults.MultiUploadThreadNum)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  V", i+1, t.Errorf, tCase.mulitNum,
				serverConfigFileProvider.cfg.Defaults.MultiUploadThreadNum)
		}
		if _, ok := strconv.Atoi(tCase.partSize); ok != nil {
			util.ExpectEqual("config.go ConfigInteractive  VI", i+1, t.Errorf, "",
				serverConfigFileProvider.cfg.Defaults.MultiUploadPartSize)
		} else {
			util.ExpectEqual("config.go ConfigInteractive  VII", i+1, t.Errorf, tCase.partSize,
				serverConfigFileProvider.cfg.Defaults.MultiUploadPartSize)
		}
	}
	os.Stdin = temp
}
