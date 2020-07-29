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
	fileServerProvider11 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				Domain:                   "bj.bcebos.com",
				Region:                   "bj",
				AutoSwitchDomain:         "yes",
				BreakpointFileExpiration: "10",
				Https:                "yes",
				MultiUploadThreadNum: "20",
				SyncProcessingNum:    "22",
			},
			Domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{
					Endpoint: "bj.bcebos.com",
				},
			},
		},
	}
	fileServerProvider12 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				AutoSwitchDomain:         "-xyes",
				BreakpointFileExpiration: "-10",
				Https:                "xdfyes",
				MultiUploadThreadNum: "-20",
				SyncProcessingNum:    "-22",
			},
			Domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{
					Endpoint: "",
				},
			},
		},
	}
	fileServerProvider13 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{},
			Domains:  map[string]*EndpointCfg{},
		},
	}
)

type getWriteConfig struct {
	provider *FileServerConfigProvider
	path     string
	isErr    bool
}

func TestWriteConfig(t *testing.T) {
	testCases := []getWriteConfig{
		getWriteConfig{
			provider: fileServerProvider11,
			path:     "./test1.cfg",
			isErr:    false,
		},
		getWriteConfig{
			provider: fileServerProvider12,
			path:     "./test1.cfg",
			isErr:    false,
		},
		getWriteConfig{
			provider: fileServerProvider13,
			path:     "./test1.cfg",
			isErr:    false,
		},
	}
	for i, tCase := range testCases {
		err := WriteConfig(tCase.path, tCase.provider.cfg)
		util.ExpectEqual("server.go WriteConfig I", i+1, t.Errorf, tCase.isErr, err != nil)
		if !tCase.isErr && err != nil {
			t.Errorf("error: %s", err)
		}
	}
}

type loadConfigType struct {
	provider *FileServerConfigProvider
	path     string
}

func TestLoadConfig(t *testing.T) {
	testCases := []loadConfigType{
		loadConfigType{
			provider: fileServerProvider11,
			path:     "./test1.cfg",
		},
		loadConfigType{
			provider: fileServerProvider12,
			path:     "./test2.cfg",
		},
		loadConfigType{
			provider: fileServerProvider13,
			path:     "./test3.cfg",
		},
	}
	for i, tCase := range testCases {
		err := WriteConfig(tCase.path, tCase.provider.cfg)
		if err != nil {
			t.Errorf("write cfg error: %s", err)
			continue
		}

		cfg := &ServerConfig{}
		if err := LoadConfig(tCase.path, cfg); err != nil {
			t.Errorf("loading cfg error: %s", err)
			continue
		}
		if len(tCase.provider.cfg.Domains) == 0 {
			if len(cfg.Domains) == 0 {
				util.ExpectEqual("server.go LoadConfig I", i+1, t.Errorf,
					tCase.provider.cfg.Defaults, cfg.Defaults)
				continue
			}
		}
		util.ExpectEqual("server.go LoadConfig I", i+1, t.Errorf, tCase.provider.cfg, cfg)
	}
}
