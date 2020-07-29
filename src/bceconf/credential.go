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

// This module provide the major operations on BOS credential configuration.

package bceconf

import (
	"fmt"
)

import (
	"utils/util"
)

const (
	DEFAULT_AK  = ""
	DEFAULT_SK  = ""
	DEFAULT_STS = ""
)

type CredentialDefaultsCfg struct {
	Ak  string // access key
	Sk  string // secret key
	Sts string // Security Token
}

type CredentialCfg struct {
	Defaults CredentialDefaultsCfg
}

type CredentialProviderInterface interface {
	GetAccessKey() (string, bool)
	GetSecretKey() (string, bool)
	GetSecurityToken() (string, bool)
}

// New credential configuration provider
func NewFileCredentialProvider(cofingPath string) (*FileCredentialProvider, error) {
	n := &FileCredentialProvider{
		configFilePath: cofingPath,
	}
	err := n.loadConfigFromFile()
	if err == nil {
		return n, nil
	}
	return nil, err
}

// Provide ak and sk for cli
type FileCredentialProvider struct {
	configFilePath string
	dirty          bool
	cfg            *CredentialCfg
}

func (f *FileCredentialProvider) loadConfigFromFile() error {
	f.cfg = &CredentialCfg{}
	if ok := util.DoesFileExist(f.configFilePath); ok {
		if err := LoadConfig(f.configFilePath, f.cfg); err != nil {
			return fmt.Errorf("load configuration error: %s", err)
		}
	}
	return nil
}

// Get Access Key.
func (f *FileCredentialProvider) GetAccessKey() (string, bool) {
	if f.cfg.Defaults.Ak != "" {
		return f.cfg.Defaults.Ak, true
	}
	return "", false
}

// Get Secret Key.
func (f *FileCredentialProvider) GetSecretKey() (string, bool) {
	if f.cfg.Defaults.Sk != "" {
		return f.cfg.Defaults.Sk, true
	}
	return "", false
}

func (f *FileCredentialProvider) GetSecurityToken() (string, bool) {
	if f.cfg.Defaults.Sts != "" {
		return f.cfg.Defaults.Sts, true
	}
	return DEFAULT_STS, false
}

// Set Access key.
func (f *FileCredentialProvider) SetAccessKey(ak string) {
	if ak != f.cfg.Defaults.Ak {
		f.cfg.Defaults.Ak = ak
		f.dirty = true
	}
}

// Set Secret key.
func (f *FileCredentialProvider) SetSecretKey(sk string) {
	if sk != f.cfg.Defaults.Sk {
		f.cfg.Defaults.Sk = sk
		f.dirty = true
	}
}

func (f *FileCredentialProvider) SetSecurityToken(sts string) {
	if sts != f.cfg.Defaults.Sts {
		f.cfg.Defaults.Sts = sts
		f.dirty = true
	}
}

// Save configuration to file
func (f *FileCredentialProvider) save() error {
	if f.configFilePath == "" {
		return fmt.Errorf("The path of credential configuration file is emtpy")
	}
	if !f.dirty {
		return nil
	}
	if err := WriteConfig(f.configFilePath, f.cfg); err == nil {
		f.dirty = false
		return nil
	} else {
		return err
	}
}

func NewDefaultCredentialProvider() (*DefaultCredentialProvider, error) {
	return &DefaultCredentialProvider{}, nil
}

type DefaultCredentialProvider struct{}

func (n *DefaultCredentialProvider) GetAccessKey() (string, bool) {
	return DEFAULT_AK, true
}

func (n *DefaultCredentialProvider) GetSecretKey() (string, bool) {
	return DEFAULT_SK, true
}

func (n *DefaultCredentialProvider) GetSecurityToken() (string, bool) {
	return DEFAULT_STS, true
}

func NewChainCredentialProvider(chain []CredentialProviderInterface) *ChainCredentialProvider {
	return &ChainCredentialProvider{chain: chain}
}

type ChainCredentialProvider struct {
	chain []CredentialProviderInterface
}

func (c *ChainCredentialProvider) GetAccessKey() (string, bool) {
	for _, provider := range c.chain {
		if val, ok := provider.GetAccessKey(); ok {
			return val, true
		}
	}
	panic("There is no access key found!")
	return "", false
}

func (c *ChainCredentialProvider) GetSecretKey() (string, bool) {
	for _, provider := range c.chain {
		if val, ok := provider.GetSecretKey(); ok {
			return val, true
		}
	}
	panic("There is no secret key found!")
	return "", false
}

func (c *ChainCredentialProvider) GetSecurityToken() (string, bool) {
	for _, provider := range c.chain {
		if val, ok := provider.GetSecurityToken(); ok {
			return val, true
		}
	}
	panic("There is no security token found!")
	return DEFAULT_STS, false
}
