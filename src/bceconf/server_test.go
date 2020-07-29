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
	"testing"
)

import (
	"utils/util"
)

var (
	defaultServerProvider, _ = NewDefaultServerConfigProvider()

	fileServerProvider1 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				Domain:                   "bj.bcebos.com",
				Region:                   "bj",
				AutoSwitchDomain:         "yes",
				BreakpointFileExpiration: "10",
				Https:                "yes",
				MultiUploadThreadNum: "20",
				MultiUploadPartSize:  "20",
				SyncProcessingNum:    "22",
			},
			Domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{
					Endpoint: "bj.bcebos.com",
				},
			},
		},
	}
	fileServerProvider2 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				AutoSwitchDomain:         "-xyes",
				BreakpointFileExpiration: "-10",
				Https:                "xdfyes",
				MultiUploadThreadNum: "-20",
				MultiUploadPartSize:  "-20",
				SyncProcessingNum:    "-22",
			},
			Domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{
					Endpoint: "",
				},
			},
		},
	}
	fileServerProvider3 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{},
			Domains:  map[string]*EndpointCfg{},
		},
	}
	fileServerProvider4 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				Domain:                   "bj.bcebos.com",
				Region:                   "bj",
				AutoSwitchDomain:         "yes",
				BreakpointFileExpiration: "10",
				Https:                "yes",
				MultiUploadPartSize:  "5.1",
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
	fileServerProvider5 = &FileServerConfigProvider{
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
	fileServerProvider6 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{},
			Domains:  map[string]*EndpointCfg{},
		},
	}
)

func init() {
	err := util.TryMkdir("./test_file/")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file")
		os.Exit(2)
	}
	fd, err := os.Create("./test_file/test.cfg")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file/test.cfg")
		os.Exit(2)
	}
	cfgVal := "[Defaults]\nDomain = bj.bcebos.com\nRegion = bj\nAutoSwitchDomain = yes\n" +
		"BreakpointFileExpiration = 10\nHttps = no\nMultiUploadThreadNum = 5\n" +
		"SyncProcessingNum = 2\n"
	fmt.Fprintf(fd, cfgVal)
	fd.Close()

	fd, err = os.Create("./test_file/test_wrong.cfg")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file/test_wrong.cfg")
		os.Exit(2)
	}
	cfgVal = "[Defaults]\nDomains = bj.bcebos.com\nRegions = bj\nAutoSwitchDomain = yes\n" +
		"BreakpointFileExpiration = 10\nHttps = no\nMultiUploadThreadNum = aaa\n" +
		"SyncProcessingNum = 2\n"
	fmt.Fprintf(fd, cfgVal)
	fd.Close()

	fd, err = os.Create("./test_file/test_wrong1.cfg")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file/test_wrong1.cfg")
		os.Exit(2)
	}
	cfgVal = "[Defaults]\nDomain = bj.bcebos.com\nRegion = bj\nAutoSwitchDomain = yes\n" +
		"BreakpointFileExpiration = -10\nHttps = no\nMultiUploadThreadNum = 5\n" +
		"SyncProcessingNum = 2\n"
	fmt.Fprintf(fd, cfgVal)
	fd.Close()

}

type serverCheckConfigType struct {
	cfg   *ServerConfig
	isErr bool
	err   error
}

func TestServerCheckConfig(t *testing.T) {
	testCases := []serverCheckConfigType{
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					BreakpointFileExpiration: "aaa",
				},
			},
			isErr: true,
			err: fmt.Errorf("BreakpointFileExpiration must be integer, and equal" +
				"or greater than -1"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					BreakpointFileExpiration: "-2231231232131232132321312",
				},
			},
			isErr: true,
			err: fmt.Errorf("BreakpointFileExpiration must be integer, and equal" +
				"or greater than -1"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					MultiUploadThreadNum: "-2231231232131232132321312",
				},
			},
			isErr: true,
			err:   fmt.Errorf("Multi upload thread number must be integer and  greater than zero!"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					MultiUploadThreadNum: "-22",
				},
			},
			isErr: true,
			err:   fmt.Errorf("Multi upload thread number must be integer and  greater than zero!"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					SyncProcessingNum: "-22",
				},
			},
			isErr: true,
			err:   fmt.Errorf("the number of sync processing must greater than zero!"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					SyncProcessingNum: "-20000000000000000000000000000000000000002",
				},
			},
			isErr: true,
			err:   fmt.Errorf("the number of sync processing must greater than zero!"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					SyncProcessingNum: "-0",
				},
			},
			isErr: true,
			err:   fmt.Errorf("the number of sync processing must greater than zero!"),
		},
		serverCheckConfigType{
			cfg: &ServerConfig{
				Defaults: ServerDefaultsCfg{
					SyncProcessingNum: "10",
				},
			},
			isErr: false,
		},
	}
	for i, tCase := range testCases {
		err := checkConfig(tCase.cfg)
		if tCase.isErr == true {
			util.ExpectEqual("server.go checkConfig", i+1, t.Errorf, tCase.err, err)
		} else {
			util.ExpectEqual("server.go checkConfig", i+1, t.Errorf, tCase.isErr, err != nil)
			if tCase.isErr == false && err != nil {
				t.Logf("id %d err: %d", i+1, err)
			}
		}
	}
}

type newFileServerConfigProviderType struct {
	path  string
	isErr bool
}

func TestNewFileServerConfigProvider(t *testing.T) {
	testCases := []newFileServerConfigProviderType{
		newFileServerConfigProviderType{
			path:  "./xxxx.cfg",
			isErr: false,
		},
		newFileServerConfigProviderType{},
		newFileServerConfigProviderType{
			path:  "./test_file/test.cfg",
			isErr: false,
		},
		newFileServerConfigProviderType{
			path:  "./test_file/test_wrong.cfg",
			isErr: true,
		},
	}
	for i, tCase := range testCases {
		ret, err := NewFileServerConfigProvider(tCase.path)
		util.ExpectEqual("server.go NewFileServerConfigProvider I", i+1, t.Errorf, tCase.isErr,
			err != nil)
		if tCase.isErr == false && err != nil {
			t.Logf("id %d error: %s\n", i+1, err)
		}
		if err == nil && ret == nil {
			t.Errorf("server.go NewFileServerConfigProvider II id %d, err is nil but ret ==nil",
				i+1)
		}
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	testCases := []newFileServerConfigProviderType{
		newFileServerConfigProviderType{
			path:  "./xxxx.cfg",
			isErr: false,
		},
		newFileServerConfigProviderType{},
		newFileServerConfigProviderType{
			path:  "./test_file/test.cfg",
			isErr: false,
		},
		newFileServerConfigProviderType{
			path:  "./test_file/test_wrong.cfg",
			isErr: true,
		},
		newFileServerConfigProviderType{
			path:  "./test_file/test_wrong1.cfg",
			isErr: true,
		},
	}

	for i, tCase := range testCases {
		n := &FileServerConfigProvider{
			configFilePath: tCase.path,
		}
		err := n.loadConfigFromFile()

		util.ExpectEqual("server.go TestLoadConfigFromFile I", i+1, t.Errorf, tCase.isErr,
			err != nil)
		if err == nil {
			t.Logf("config of %s is %v", tCase.path, n.cfg.Defaults)
		}
	}
}

type getDomainByRegionType struct {
	provider *FileServerConfigProvider
	region   string
	domain   string
	isSuc    bool
}

func TestGetDomainByRegion(t *testing.T) {
	testCases := []getDomainByRegionType{
		getDomainByRegionType{
			provider: fileServerProvider1,
			region:   "bj",
			domain:   "bj.bcebos.com",
			isSuc:    true,
		},
		getDomainByRegionType{
			provider: fileServerProvider2,
			region:   "bj",
			domain:   "bj.bcebos.com",
			isSuc:    true,
		},
		getDomainByRegionType{
			provider: fileServerProvider2,
			region:   "gz",
			domain:   "gz.bcebos.com",
			isSuc:    true,
		},
		getDomainByRegionType{
			provider: fileServerProvider2,
			region:   "sc",
			isSuc:    false,
		},
		getDomainByRegionType{
			provider: fileServerProvider2,
			region:   "",
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetDomainByRegion(tCase.region)
		util.ExpectEqual("server.go TestGetDomainByRegionType I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go TestGetDomainByRegionType II", i+1, t.Errorf, tCase.domain,
				ret)
		}
	}
}

type getDomainType struct {
	provider *FileServerConfigProvider
	domain   string
	isSuc    bool
}

func TestGetDomain(t *testing.T) {
	testCases := []getDomainType{
		getDomainType{
			provider: fileServerProvider1,
			domain:   "bj.bcebos.com",
			isSuc:    true,
		},
		getDomainType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetDomain()
		util.ExpectEqual("server.go GetDomain I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetDomain II", i+1, t.Errorf, tCase.domain, ret)
		}
	}
}

type getRegionType struct {
	provider *FileServerConfigProvider
	region   string
	isSuc    bool
}

func TestGetRegion(t *testing.T) {
	testCases := []getRegionType{
		getRegionType{
			provider: fileServerProvider1,
			region:   "bj",
			isSuc:    true,
		},
		getRegionType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetRegion()
		util.ExpectEqual("server.go GetRegion I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetRegion II", i+1, t.Errorf, tCase.region, ret)
		}
	}
}

type getUseAutoSwitchDomainType struct {
	provider *FileServerConfigProvider
	ret      bool
	isSuc    bool
}

func TestGetUseAutoSwitchDomain(t *testing.T) {
	testCases := []getUseAutoSwitchDomainType{
		getUseAutoSwitchDomainType{
			provider: fileServerProvider1,
			ret:      true,
			isSuc:    true,
		},
		getUseAutoSwitchDomainType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getUseAutoSwitchDomainType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetUseAutoSwitchDomain()
		util.ExpectEqual("server.go GetRegion I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetRegion II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getBreakpointFileExpirationType struct {
	provider *FileServerConfigProvider
	ret      int
	isSuc    bool
}

func TestBreakpointFileExpiration(t *testing.T) {
	testCases := []getBreakpointFileExpirationType{
		getBreakpointFileExpirationType{
			provider: fileServerProvider1,
			ret:      10,
			isSuc:    true,
		},
		getBreakpointFileExpirationType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getBreakpointFileExpirationType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetBreakpointFileExpiration()
		util.ExpectEqual("server.go BreakpointFileExpiration I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go BreakpointFileExpiration II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getUseHttpsProtocolType struct {
	provider *FileServerConfigProvider
	ret      bool
	isSuc    bool
}

func TestUseHttpsProtocol(t *testing.T) {
	testCases := []getUseHttpsProtocolType{
		getUseHttpsProtocolType{
			provider: fileServerProvider1,
			ret:      true,
			isSuc:    true,
		},
		getUseHttpsProtocolType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getUseHttpsProtocolType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetUseHttpsProtocol()
		util.ExpectEqual("server.go GetUseHttpsProtocol I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetUseHttpsProtocol II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getMultiUploadThreadNumType struct {
	provider *FileServerConfigProvider
	ret      int
	isSuc    bool
}

func TestMultiUploadThreadNum(t *testing.T) {
	testCases := []getMultiUploadThreadNumType{
		getMultiUploadThreadNumType{
			provider: fileServerProvider1,
			ret:      20,
			isSuc:    true,
		},
		getMultiUploadThreadNumType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getMultiUploadThreadNumType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetMultiUploadThreadNum()
		util.ExpectEqual("server.go GetMultiUploadThreadNum I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetMultiUploadThreadNum II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getMultiUploadPartSizeType struct {
	provider *FileServerConfigProvider
	ret      int
	isSuc    bool
}

func TestMultiUploadPartSize(t *testing.T) {
	testCases := []getMultiUploadPartSizeType{
		getMultiUploadPartSizeType{
			provider: fileServerProvider1,
			ret:      20,
			isSuc:    true,
		},
		getMultiUploadPartSizeType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getMultiUploadPartSizeType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
		getMultiUploadPartSizeType{
			provider: fileServerProvider4,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetMultiUploadPartSize()
		util.ExpectEqual("server.go GetMultiUploadPartSize I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetMultiUploadPartSize II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type getSyncProcessingNumType struct {
	provider *FileServerConfigProvider
	ret      int
	isSuc    bool
}

func TestGetSyncProcessingNum(t *testing.T) {
	testCases := []getSyncProcessingNumType{
		getSyncProcessingNumType{
			provider: fileServerProvider1,
			ret:      22,
			isSuc:    true,
		},
		getSyncProcessingNumType{
			provider: fileServerProvider2,
			isSuc:    false,
		},
		getSyncProcessingNumType{
			provider: fileServerProvider3,
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetSyncProcessingNum()
		util.ExpectEqual("server.go GetSyncProcessingNum I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("server.go GetSyncProcessingNum II", i+1, t.Errorf, tCase.ret, ret)
		}
	}
}

type setDomainType struct {
	provider *FileServerConfigProvider
	domain   string
	dirty    bool
}

func TestSetDomain(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setDomainType{
		setDomainType{
			provider: fileServerProvider4,
			domain:   "bj.bcebos.com",
			dirty:    false,
		},
		setDomainType{
			provider: fileServerProvider4,
			domain:   "ga.bcebos.com",
			dirty:    true,
		},
		setDomainType{
			provider: fileServerProvider6,
			domain:   "ga.bcebos.com",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.SetDomain(tCase.domain)
		util.ExpectEqual("server.go SetDomain I", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
		util.ExpectEqual("server.go SetDomain II", i+1, t.Errorf, tCase.domain,
			tCase.provider.cfg.Defaults.Domain)
	}
}

type setDomainsType struct {
	provider *FileServerConfigProvider
	domains  map[string]*EndpointCfg
	dirty    bool
}

func TestSetDomains(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setDomainsType{
		setDomainsType{
			provider: fileServerProvider4,
			domains: map[string]*EndpointCfg{
				"bj": &EndpointCfg{Endpoint: "bj.bcebos.com"},
				"gz": &EndpointCfg{Endpoint: "gz.bcebos.com"},
				"yq": &EndpointCfg{Endpoint: "yq.bcebos.com"},
			},
			dirty: true,
		},
		setDomainsType{
			provider: fileServerProvider4,
			domains:  map[string]*EndpointCfg{},
			dirty:    true,
		},
		setDomainsType{
			provider: fileServerProvider6,
			domains:  nil,
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.SetDomains(tCase.domains)
		util.ExpectEqual("server.go SetDomains I", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
		util.ExpectEqual("server.go SetDomains II", i+1, t.Errorf, tCase.domains,
			tCase.provider.cfg.Domains)
	}
}

type delDomainInDomainsType struct {
	provider *FileServerConfigProvider
	region   string
	dirty    bool
}

func TestDelDomains(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []delDomainInDomainsType{
		delDomainInDomainsType{
			provider: fileServerProvider4,
			region:   "bj",
			dirty:    true,
		},
		delDomainInDomainsType{
			provider: fileServerProvider4,
			region:   "23xx",
			dirty:    true,
		},
		delDomainInDomainsType{
			provider: fileServerProvider6,
			region:   "gz",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.DelDomainInDomains(tCase.region)
		util.ExpectEqual("server.go SetDomains I", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
		if ret, ok := tCase.provider.cfg.Domains[tCase.region]; ok {
			t.Errorf("delete %s from domains failed, domain '%s'", tCase.region, ret)
		}
	}
}

type insertDomainIntoDomainsType struct {
	provider *FileServerConfigProvider
	region   string
	domain   string
	isSucc   bool
	dirty    bool
}

func TestInsertDomainIntoDomains(t *testing.T) {
	testCases := []insertDomainIntoDomainsType{
		insertDomainIntoDomainsType{
			provider: &FileServerConfigProvider{
				cfg: &ServerConfig{
					Domains: map[string]*EndpointCfg{
						"bj": &EndpointCfg{
							Endpoint: "bj.bcebos.com",
						},
					},
				},
			},
			region: "bj",
			domain: "bj.bcebos.com",
			isSucc: true,
			dirty:  false,
		},
		insertDomainIntoDomainsType{
			provider: &FileServerConfigProvider{
				cfg: &ServerConfig{},
			},
			region: "bj",
			domain: "bj.bcebos.com",
			isSucc: true,
			dirty:  true,
		},
		insertDomainIntoDomainsType{
			provider: &FileServerConfigProvider{
				cfg: &ServerConfig{},
			},
			region: "bj",
			domain: "gz.bcebos.com",
			isSucc: true,
			dirty:  true,
		},
		insertDomainIntoDomainsType{
			provider: &FileServerConfigProvider{
				cfg: &ServerConfig{},
			},
			region: "",
			domain: "",
			isSucc: false,
			dirty:  false,
		},
	}
	for i, tCase := range testCases {
		ret := tCase.provider.InsertDomainIntoDomains(tCase.region, tCase.domain)
		util.ExpectEqual("server.go InsertDomainIntoDomains I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go InsertDomainIntoDomains II", i+1, t.Errorf, tCase.isSucc, ret)
		if tCase.isSucc {

			if val, ok := tCase.provider.GetDomainByRegion(tCase.region); !ok {
				t.Errorf("get domain of region %s failed", tCase.region)
				continue
			} else {
				util.ExpectEqual("server.go InsertDomainIntoDomains III", i+1, t.Errorf,
					tCase.domain, val)
			}
		}
	}
}

type setRegionType struct {
	provider *FileServerConfigProvider
	region   string
	dirty    bool
}

func TestSetRegion(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setRegionType{
		setRegionType{
			provider: fileServerProvider4,
			region:   "bj",
			dirty:    false,
		},
		setRegionType{
			provider: fileServerProvider4,
			region:   "gz",
			dirty:    true,
		},
		setRegionType{
			provider: fileServerProvider6,
			region:   "gz",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.SetRegion(tCase.region)
		util.ExpectEqual("server.go SetRegion I", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
		util.ExpectEqual("server.go SetDomain II", i+1, t.Errorf, tCase.region,
			tCase.provider.cfg.Defaults.Region)
	}
}

type setSetUseAutoSwitchDomain struct {
	provider *FileServerConfigProvider
	use      string
	dirty    bool
	ret      bool
}

func TestSetUseAutoSwitchDomain(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "yes",
			dirty:    false,
			ret:      true,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
			ret:      false,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider6,
			use:      "xxx",
			dirty:    true,
			ret:      false,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetUseAutoSwitchDomain(tCase.use)
		util.ExpectEqual("server.go SetUseAutoSwitchDomain I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetUseAutoSwitchDomain II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.AutoSwitchDomain)
		use, _ := tCase.provider.GetUseAutoSwitchDomain()
		util.ExpectEqual("server.go SetUseAutoSwitchDomain III", i+1, t.Errorf, tCase.ret, use)
	}
}

func TestSetBreakpointFileExpiration(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "10",
			dirty:    false,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetBreakpointFileExpiration(tCase.use)
		util.ExpectEqual("server.go SetBreakpointFileExpiration I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetBreakpointFileExpiration II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.BreakpointFileExpiration)
	}
}

func TestSetUseHttpsProtocol(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "yes",
			dirty:    false,
			ret:      true,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
			ret:      false,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider6,
			use:      "xxx",
			dirty:    true,
			ret:      false,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetUseHttpsProtocol(tCase.use)
		util.ExpectEqual("server.go SetUseHttpsProtocol I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetUseHttpsProtocol II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.Https)
	}
}

func TestSetMultiUploadThreadNum(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "20",
			dirty:    false,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetMultiUploadThreadNum(tCase.use)
		util.ExpectEqual("server.go SetMultiUploadThreadNum I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetMultiUploadThreadNum II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.MultiUploadThreadNum)
	}
}

func TestSetMultiUploadPartSize(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "20",
			dirty:    true,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetMultiUploadPartSize(tCase.use)
		util.ExpectEqual("server.go SetMultiUploadPartSize I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetMultiUploadPartSize II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.MultiUploadPartSize)
	}
}

func TestSetSyncProcessingNum(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []setSetUseAutoSwitchDomain{
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "22",
			dirty:    false,
		},
		setSetUseAutoSwitchDomain{
			provider: fileServerProvider4,
			use:      "no",
			dirty:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		tCase.provider.SetSyncProcessingNum(tCase.use)
		util.ExpectEqual("server.go SetSyncProcessingNum I", i+1, t.Errorf, tCase.dirty,
			tCase.provider.dirty)
		util.ExpectEqual("server.go SetSyncProcessingNum II", i+1, t.Errorf, tCase.use,
			tCase.provider.cfg.Defaults.SyncProcessingNum)
	}
}

type serverSaveType struct {
	provider *FileServerConfigProvider
	path     string
	setDirty bool
	isErr    bool
	dirty    bool
}

func TestServerSave(t *testing.T) {
	fileServerProvider4.dirty = false
	fileServerProvider6.dirty = false
	testCases := []serverSaveType{
		serverSaveType{
			provider: fileServerProvider4,
			path:     "",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		serverSaveType{
			provider: fileServerProvider4,
			path:     "/root/cfg",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		serverSaveType{
			provider: fileServerProvider4,
			path:     "./test.cfg",
			setDirty: true,
			isErr:    false,
			dirty:    false,
		},
		serverSaveType{
			provider: fileServerProvider4,
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
		util.ExpectEqual("server.go save I", i+1, t.Errorf, tCase.isErr, err != nil)
		if tCase.isErr == false && err != nil {
			t.Logf("error: %s", err)
		}
		util.ExpectEqual("server.go save II", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
	}
}

type defaultGetType struct {
	input   string
	outStr  string
	outBool bool
	outInt  int
}

func TestDGetDomainByRegion(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			input:  "bj",
			outStr: "bj.bcebos.com",
		},
		defaultGetType{
			input:  "xx",
			outStr: "xx.bcebos.com",
		},
		defaultGetType{
			input:  "",
			outStr: "bj.bcebos.com",
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetDomainByRegion(tCase.input)
		util.ExpectEqual("server.go DE GetDomainByRegion I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetDomainByRegion II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestDGetRegion(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outStr: "bj",
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetRegion()
		util.ExpectEqual("server.go DE GetRegion I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetRegion II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestDDomain(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outStr: "bj.bcebos.com",
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetDomain()
		util.ExpectEqual("server.go DE GetDomain I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetDomain II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestDGetUseAutoSwitchDomain(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outBool: true,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetUseAutoSwitchDomain()
		util.ExpectEqual("server.go DE GetUseAutoSwitchDomain I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetUseAutoSwitchDomain II", i+1, t.Errorf, tCase.outBool,
			ret)
	}
}

func TestDGetBreakpointFileExpiration(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outInt: 7,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetBreakpointFileExpiration()
		util.ExpectEqual("server.go DE GetBreakpointFileExpiration I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetBreakpointFileExpiration II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

func TestDGetUseHttpsProtocol(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outBool: false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetUseHttpsProtocol()
		util.ExpectEqual("server.go DE GetUseHttpsProtocol I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetUseHttpsProtocol II", i+1, t.Errorf, tCase.outBool,
			ret)
	}
}

func TestDGetMultiUploadThreadNum(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outInt: 10,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetMultiUploadThreadNum()
		util.ExpectEqual("server.go DE GetMultiUploadThreadNum I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetMultiUploadThreadNum II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

func TestDGetMultiUploadPartSize(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outInt: 10,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetMultiUploadPartSize()
		util.ExpectEqual("server.go DE GetMultiUploadPartSize I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetMultiUploadPartSize II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

func TestDGetSyncProcessingNum(t *testing.T) {
	testCases := []defaultGetType{
		defaultGetType{
			outInt: 10,
		},
	}
	for i, tCase := range testCases {
		ret, ok := defaultServerProvider.GetSyncProcessingNum()
		util.ExpectEqual("server.go DE GetSyncProcessingNum I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go DE GetSyncProcessingNum II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

var (
	fileServerProvider7 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{
				Domain:                   "gz.bcebos.com",
				Region:                   "gz",
				AutoSwitchDomain:         "yes",
				BreakpointFileExpiration: "10",
				Https:                "yes",
				MultiUploadThreadNum: "20",
				SyncProcessingNum:    "22",
			},
			Domains: map[string]*EndpointCfg{
				"gz": &EndpointCfg{
					Endpoint: "gz.bcebos.com",
				},
			},
		},
	}
	fileServerProvider8 = &FileServerConfigProvider{
		cfg: &ServerConfig{
			Defaults: ServerDefaultsCfg{},
			Domains:  map[string]*EndpointCfg{},
		},
	}
	chainServerProvider1 = NewChainServerConfigProvider([]ServerConfigProviderInterface{
		fileServerProvider7, defaultServerProvider})
	chainServerProvider2 = NewChainServerConfigProvider([]ServerConfigProviderInterface{
		fileServerProvider8, defaultServerProvider})
)

type chainServerType struct {
	provider *ChainServerConfigProvider
	input    string
	outStr   string
	outBool  bool
	outInt   int
}

func TestCGetDomainByRegion(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			input:    "bj",
			outStr:   "bj.bcebos.com",
		},
		chainServerType{
			provider: chainServerProvider2,
			input:    "xx",
			outStr:   "xx.bcebos.com",
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetDomainByRegion(tCase.input)
		util.ExpectEqual("server.go ch GetDomainByRegion I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetDomainByRegion II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestCGetDomain(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outStr:   "gz.bcebos.com",
		},
		chainServerType{
			provider: chainServerProvider2,
			outStr:   "bj.bcebos.com",
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetDomain()
		util.ExpectEqual("server.go ch GetDomain I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetDomain II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestCGetRegion(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outStr:   "gz",
		},
		chainServerType{
			provider: chainServerProvider2,
			outStr:   "bj",
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetRegion()
		util.ExpectEqual("server.go ch GetRegion I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetRegion II", i+1, t.Errorf, tCase.outStr, ret)
	}
}

func TestCGetUseAutoSwitchDomain(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outBool:  true,
		},
		chainServerType{
			provider: chainServerProvider2,
			outBool:  true,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetUseAutoSwitchDomain()
		util.ExpectEqual("server.go ch GetUseAutoSwitchDomain I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetUseAutoSwitchDomain II", i+1, t.Errorf, tCase.outBool,
			ret)
	}
}

func TestCGetBreakpointFileExpiration(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outInt:   10,
		},
		chainServerType{
			provider: chainServerProvider2,
			outInt:   7,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetBreakpointFileExpiration()
		util.ExpectEqual("server.go ch GetBreakpointFileExpiration I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetBreakpointFileExpiration II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

func TestCGetUseHttpsProtocol(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outBool:  true,
		},
		chainServerType{
			provider: chainServerProvider2,
			outBool:  false,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetUseHttpsProtocol()
		util.ExpectEqual("server.go ch GetUseHttpsProtocol I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetUseHttpsProtocol II", i+1, t.Errorf, tCase.outBool,
			ret)
	}
}

func TestCGetMultiUploadThreadNum(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outInt:   20,
		},
		chainServerType{
			provider: chainServerProvider2,
			outInt:   10,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetMultiUploadThreadNum()
		util.ExpectEqual("server.go ch GetMultiUploadThreadNum I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetMultiUploadThreadNum II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}

func TestCGetSyncProcessingNum(t *testing.T) {
	testCases := []chainServerType{
		chainServerType{
			provider: chainServerProvider1,
			outInt:   22,
		},
		chainServerType{
			provider: chainServerProvider2,
			outInt:   10,
		},
	}
	for i, tCase := range testCases {
		ret, ok := tCase.provider.GetSyncProcessingNum()
		util.ExpectEqual("server.go ch GetSyncProcessingNum I", i+1, t.Errorf, true, ok)
		util.ExpectEqual("server.go ch GetSyncProcessingNum II", i+1, t.Errorf,
			tCase.outInt, ret)
	}
}
