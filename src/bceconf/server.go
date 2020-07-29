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

// This file provide basic configuration of BCE Server

package bceconf

import (
	"fmt"
	"strconv"
)

import (
	"utils/util"
)

const (
	SERVER_SECTION_NAME                    = "defaults"
	DOMAIN_OPTION_NAME                     = "domain"
	REGION_OPTION_NAME                     = "region"
	USE_AUTO_SWITCH_DOMAIN_OPTION_NAME     = "auto_switch_domain"
	BREAKPIONT_FILE_EXPIRATION_OPTION_NAME = "breakpoint_file_expiration"
	USE_HTTPS_OPTION_NAME                  = "https"
	MULTI_UPLOAD_THREAD_NUM_NAME           = "multi_upload_thread_num"
	DEFAULT_DOMAIN_SUFFIX                  = ".bcebos.com"
	DEFAULT_REGION                         = "bj"
	DEFAULT_USE_AUTO_SWITCH_DOMAIN         = "yes"
	DEFAULT_BREAKPIONT_FILE_EXPIRATION     = "7"
	DEFAULT_USE_HTTPS_PROTOCOL             = "no"
	DEFAULT_MULTI_UPLOAD_THREAD_NUM        = "10"
	DEFAULT_MULTI_UPLOAD_PART_SIZE         = "10"
	DEFAULT_SYNC_PROCESSING_NUM            = "10"
	WILL_USE_AUTO_SWTICH_DOMAIN            = "yes"
	DOMAINS_SECTION_NAME                   = "domains"
)

var (
	AOLLOWED_CONFIRM_OPTIONS = map[string]bool{
		"y":   true,
		"yes": true,
		"Yes": true,
		"YES": true,
		"n":   false,
		"no":  false,
		"No":  false,
		"NO":  false,
	}
	BOOL_TO_STRING = map[bool]string{
		true:  "yes",
		false: "no",
	}

	DEFAULT_DOMAINS = map[string]string{
		"bj":   "bj.bcebos.com",
		"gz":   "gz.bcebos.com",
		"su":   "su.bcebos.com",
		"hk02": "hk-2.bcebos.com",
		"hkg":  "hkg.bcebos.com",
		"yq":   "bos.yq.baidubce.com",
	}
)

// Store the default configuration.
// The default value of int is zero, therefore each parameter is special as string .
type ServerDefaultsCfg struct {
	Domain                   string
	Region                   string
	AutoSwitchDomain         string
	BreakpointFileExpiration string
	Https                    string
	MultiUploadThreadNum     string
	SyncProcessingNum        string
	MultiUploadPartSize      string
}

// Store region => domain
type EndpointCfg struct {
	Endpoint string
}

type ServerConfig struct {
	Defaults ServerDefaultsCfg
	Domains  map[string]*EndpointCfg
}

func checkConfig(cfg *ServerConfig) error {
	if cfg == nil {
		return nil
	}
	if cfg.Defaults.BreakpointFileExpiration != "" {
		val, ok := strconv.Atoi(cfg.Defaults.BreakpointFileExpiration)
		if ok != nil || val < -1 {
			return fmt.Errorf("BreakpointFileExpiration must be integer, and equal" +
				"or greater than -1")
		}
	}
	if cfg.Defaults.MultiUploadThreadNum != "" {
		val, ok := strconv.Atoi(cfg.Defaults.MultiUploadThreadNum)
		if ok != nil || val < 1 {
			return fmt.Errorf("Multi upload thread number must be integer and  greater than zero!")
		}
	}
	if cfg.Defaults.SyncProcessingNum != "" {
		val, ok := strconv.Atoi(cfg.Defaults.SyncProcessingNum)
		if ok != nil || val < 1 {
			return fmt.Errorf("the number of sync processing must greater than zero!")
		}
	}
	if cfg.Defaults.MultiUploadPartSize != "" {
		val, ok := strconv.Atoi(cfg.Defaults.MultiUploadPartSize)
		if ok != nil || val < 1 || val%1 != 0 {
			return fmt.Errorf("part size must greater than zero!")
		}
	}
	return nil
}

type ServerConfigProviderInterface interface {
	GetDomain() (string, bool)
	GetDomainByRegion(string) (string, bool)
	GetRegion() (string, bool)
	GetUseAutoSwitchDomain() (bool, bool)
	GetBreakpointFileExpiration() (int, bool)
	GetUseHttpsProtocol() (bool, bool)
	GetMultiUploadThreadNum() (int64, bool)
	GetSyncProcessingNum() (int, bool)
	GetMultiUploadPartSize() (int64, bool)
}

// New file configuration provider
func NewFileServerConfigProvider(cofingPath string) (*FileServerConfigProvider, error) {
	n := &FileServerConfigProvider{
		configFilePath: cofingPath,
	}
	err := n.loadConfigFromFile()
	if err == nil {
		return n, nil
	}
	return nil, err
}

// Read server configuration from a file
type FileServerConfigProvider struct {
	configFilePath string
	dirty          bool
	cfg            *ServerConfig
}

// When configuration file exist, loading configuration from a file
func (f *FileServerConfigProvider) loadConfigFromFile() error {
	f.cfg = &ServerConfig{}
	if ok := util.DoesFileExist(f.configFilePath); ok {
		if err := LoadConfig(f.configFilePath, f.cfg); err != nil {
			return fmt.Errorf("load configuration error: %s", err)
		}
		if err := checkConfig(f.cfg); err != nil {
			return err
		}
	}
	return nil
}

// Get server domain by region
// return: The domian of region
func (f *FileServerConfigProvider) GetDomainByRegion(region string) (string, bool) {
	if region == "" {
		return "", false
	}
	if len(f.cfg.Domains) > 0 {
		domainInfo, ok := f.cfg.Domains[region]
		if ok && domainInfo.Endpoint != "" {
			return domainInfo.Endpoint, true
		}
	}
	domain, ok := DEFAULT_DOMAINS[region]
	if ok && domain != "" {
		return domain, true
	}
	return "", false
}

// Get server domain address
func (f *FileServerConfigProvider) GetDomain() (string, bool) {
	if f.cfg.Defaults.Domain != "" {
		return f.cfg.Defaults.Domain, true
	}
	return "", false
}

// Return server region.
func (f *FileServerConfigProvider) GetRegion() (string, bool) {
	if f.cfg.Defaults.Region != "" {
		return f.cfg.Defaults.Region, true
	}
	return "", false
}

//  return use auto siwitch domain ('yes' or 'no' or empty)
func (f *FileServerConfigProvider) GetUseAutoSwitchDomain() (bool, bool) {
	if f.cfg.Defaults.AutoSwitchDomain != "" {
		if val, ok := AOLLOWED_CONFIRM_OPTIONS[f.cfg.Defaults.AutoSwitchDomain]; ok {
			return val, true
		}
	}
	return false, false
}

// return: Breakpoint file expiration
func (f *FileServerConfigProvider) GetBreakpointFileExpiration() (int, bool) {
	if f.cfg.Defaults.BreakpointFileExpiration != "" {
		if val, ok := strconv.Atoi(f.cfg.Defaults.BreakpointFileExpiration); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get wheather use https
// RETURN: true or false
func (f *FileServerConfigProvider) GetUseHttpsProtocol() (bool, bool) {
	if f.cfg.Defaults.Https != "" {
		if val, ok := AOLLOWED_CONFIRM_OPTIONS[f.cfg.Defaults.Https]; ok {
			return val, true
		}
	}
	return false, false
}

// return: Server is multi upload thread num
func (f *FileServerConfigProvider) GetMultiUploadThreadNum() (int64, bool) {
	if f.cfg.Defaults.MultiUploadThreadNum != "" {
		if val, ok := strconv.ParseInt(f.cfg.Defaults.MultiUploadThreadNum, 10, 64); ok == nil {
			if val > 0 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get sync processing num number
func (f *FileServerConfigProvider) GetSyncProcessingNum() (int, bool) {
	if f.cfg.Defaults.SyncProcessingNum != "" {
		if val, ok := strconv.Atoi(f.cfg.Defaults.SyncProcessingNum); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get sync processing num number
func (f *FileServerConfigProvider) GetMultiUploadPartSize() (int64, bool) {
	if f.cfg.Defaults.MultiUploadPartSize != "" {
		if val, ok := strconv.ParseInt(f.cfg.Defaults.MultiUploadPartSize, 10, 64); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// param domain: Set server domain address
// domain can be empty
func (f *FileServerConfigProvider) SetDomain(domain string) {
	if f.cfg.Defaults.Domain != domain {
		f.cfg.Defaults.Domain = domain
		f.dirty = true
	}
}

// Set server domains address.
// domains is  {region1: domain1, region2: domain2 }
// domains can be empty
func (f *FileServerConfigProvider) SetDomains(domains map[string]*EndpointCfg) {
	f.cfg.Domains = domains
	f.dirty = true
}

// Delete a domain from domains, return nothing
func (f *FileServerConfigProvider) DelDomainInDomains(region string) {
	if region != "" {
		delete(f.cfg.Domains, region)
		f.dirty = true
	}
}

// param region: Set server domain by region
// return:false or true
func (f *FileServerConfigProvider) InsertDomainIntoDomains(region, domain string) bool {
	if region == "" || domain == "" {
		return false
	}
	val, ok := f.cfg.Domains[region]
	if ok && val != nil && val.Endpoint == domain {
		return true
	}
	if f.cfg.Domains == nil {
		f.cfg.Domains = make(map[string]*EndpointCfg)
	}
	f.cfg.Domains[region] = &EndpointCfg{Endpoint: domain}
	f.dirty = true
	return true
}

// Set server region
func (f *FileServerConfigProvider) SetRegion(region string) bool {
	if f.cfg.Defaults.Region != region {
		f.cfg.Defaults.Region = region
		f.dirty = true
	}
	return true
}

// Set use auto siwitch domain ("yes" or "no")
func (f *FileServerConfigProvider) SetUseAutoSwitchDomain(useAutoSwitchDomain string) bool {
	if f.cfg.Defaults.AutoSwitchDomain != useAutoSwitchDomain {
		f.cfg.Defaults.AutoSwitchDomain = useAutoSwitchDomain
		f.dirty = true
	}
	return true
}

// param breakpoint_file_expiration: Set breakpoint file expiration
// num can be empty
func (f *FileServerConfigProvider) SetBreakpointFileExpiration(num string) bool {
	if f.cfg.Defaults.BreakpointFileExpiration != num {
		f.cfg.Defaults.BreakpointFileExpiration = num
		f.dirty = true
	}
	return true
}

// set use https protocol
func (f *FileServerConfigProvider) SetUseHttpsProtocol(useHttpsProtocol string) bool {
	if f.cfg.Defaults.Https != useHttpsProtocol {
		f.cfg.Defaults.Https = useHttpsProtocol
		f.dirty = true
	}
	return true
}

// set multi uplaod thread number
func (f *FileServerConfigProvider) SetMultiUploadThreadNum(multiUploadThreadNum string) bool {
	if f.cfg.Defaults.MultiUploadThreadNum != multiUploadThreadNum {
		f.cfg.Defaults.MultiUploadThreadNum = multiUploadThreadNum
		f.dirty = true
	}
	return true
}

// set sync processing number
func (f *FileServerConfigProvider) SetSyncProcessingNum(syncProcessingNum string) bool {
	if f.cfg.Defaults.SyncProcessingNum != syncProcessingNum {
		f.cfg.Defaults.SyncProcessingNum = syncProcessingNum
		f.dirty = true
	}
	return true
}

// set mulit upload part size
func (f *FileServerConfigProvider) SetMultiUploadPartSize(multiUploadPartSize string) bool {
	if multiUploadPartSize != f.cfg.Defaults.MultiUploadPartSize {
		f.cfg.Defaults.MultiUploadPartSize = multiUploadPartSize
		f.dirty = true
	}
	return true
}

// Save configuration into file
func (f *FileServerConfigProvider) save() error {
	if f.configFilePath == "" {
		return fmt.Errorf("The path of configuration file is emtpy")
	} else if !f.dirty {
		return nil
	}
	if err := WriteConfig(f.configFilePath, f.cfg); err == nil {
		f.dirty = false
		return nil
	} else {
		return err
	}
}

func NewDefaultServerConfigProvider() (*DefaultServerConfigProvider, error) {
	return &DefaultServerConfigProvider{}, nil
}

// Provide default value for serve configuration
type DefaultServerConfigProvider struct{}

// Get default domain
// default domain is region + ".bcebos.com"
func (d *DefaultServerConfigProvider) GetDomain() (string, bool) {
	return DEFAULT_REGION + DEFAULT_DOMAIN_SUFFIX, true
}

// Get default domain of region
// if region is empty, return defalut-region + ".bcebos.com"
// else find the domian of region in DEFAULT_DOMAINS
func (d *DefaultServerConfigProvider) GetDomainByRegion(region string) (string, bool) {
	if region != "" {
		domain, ok := DEFAULT_DOMAINS[region]
		if ok && domain != "" {
			return domain, true
		} else {
			return region + DEFAULT_DOMAIN_SUFFIX, true
		}
	}
	return DEFAULT_REGION + DEFAULT_DOMAIN_SUFFIX, true
}

// Get default region
func (d *DefaultServerConfigProvider) GetRegion() (string, bool) {
	return DEFAULT_REGION, true
}

// Get wheather use auto siwitch domain
func (d *DefaultServerConfigProvider) GetUseAutoSwitchDomain() (bool, bool) {
	val, ok := AOLLOWED_CONFIRM_OPTIONS[DEFAULT_USE_AUTO_SWITCH_DOMAIN]
	if ok {
		return val, true
	}
	return false, false
}

// Get breakpoint file expiration
func (d *DefaultServerConfigProvider) GetBreakpointFileExpiration() (int, bool) {
	if DEFAULT_BREAKPIONT_FILE_EXPIRATION != "" {
		if val, ok := strconv.Atoi(DEFAULT_BREAKPIONT_FILE_EXPIRATION); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get wheather use https protocol
func (d *DefaultServerConfigProvider) GetUseHttpsProtocol() (bool, bool) {
	val, ok := AOLLOWED_CONFIRM_OPTIONS[DEFAULT_USE_HTTPS_PROTOCOL]
	if ok {
		return val, true
	}
	return false, false
}

// Get sever multi upload thread number
func (d *DefaultServerConfigProvider) GetMultiUploadThreadNum() (int64, bool) {
	if DEFAULT_MULTI_UPLOAD_THREAD_NUM != "" {
		if val, ok := strconv.ParseInt(DEFAULT_MULTI_UPLOAD_THREAD_NUM, 10, 64); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get sync processing num number
func (d *DefaultServerConfigProvider) GetMultiUploadPartSize() (int64, bool) {
	if DEFAULT_MULTI_UPLOAD_PART_SIZE != "" {
		if val, ok := strconv.ParseInt(DEFAULT_MULTI_UPLOAD_PART_SIZE, 10, 64); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

// Get sync processing num number
func (d *DefaultServerConfigProvider) GetSyncProcessingNum() (int, bool) {
	if DEFAULT_SYNC_PROCESSING_NUM != "" {
		if val, ok := strconv.Atoi(DEFAULT_SYNC_PROCESSING_NUM); ok == nil {
			if val >= -1 {
				return val, true
			}
		}
	}
	return 0, false
}

func NewChainServerConfigProvider(chain []ServerConfigProviderInterface) *ChainServerConfigProvider {
	return &ChainServerConfigProvider{chain: chain}
}

type ChainServerConfigProvider struct {
	chain []ServerConfigProviderInterface
}

// Get default domain
func (c *ChainServerConfigProvider) GetDomain() (string, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetDomain()
		if ok {
			return val, true
		}
	}
	panic("There is no domain found!")
	return "", false
}

// Get default domain of region
func (c *ChainServerConfigProvider) GetDomainByRegion(region string) (string, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetDomainByRegion(region)
		if ok {
			return val, true
		}
	}
	panic("There is no domain found!")
	return "", false
}

// Get default region
func (c *ChainServerConfigProvider) GetRegion() (string, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetRegion()
		if ok {
			return val, true
		}
	}
	panic("There is no region found!")
	return "", false
}

// Get wheather use auto siwitch domain
func (c *ChainServerConfigProvider) GetUseAutoSwitchDomain() (bool, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetUseAutoSwitchDomain()
		if ok {
			return val, true
		}
	}
	panic("There is no use_auto_switch_domain found!")
	return false, false
}

// Get breakpoint file expiration
func (c *ChainServerConfigProvider) GetBreakpointFileExpiration() (int, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetBreakpointFileExpiration()
		if ok {
			return val, true
		}
	}
	panic("There if no breakpoint file expiration found!")
	return 0, false
}

// Get wheather use https protocol
func (c *ChainServerConfigProvider) GetUseHttpsProtocol() (bool, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetUseHttpsProtocol()
		if ok {
			return val, true
		}
	}
	panic("There is no https protocol info found!")
	return false, false
}

// Get sever multi upload thread number
func (c *ChainServerConfigProvider) GetMultiUploadThreadNum() (int64, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetMultiUploadThreadNum()
		if ok {
			return val, true
		}
	}
	panic("There is no MultiUploadThreadNum found!")
	return 0, false
}

// Get sync processing num number
func (c *ChainServerConfigProvider) GetSyncProcessingNum() (int, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetSyncProcessingNum()
		if ok {
			return val, true
		}
	}
	panic("There is no info about sync processing num found!")
	return 0, false
}

// Get sever multi upload part size
func (c *ChainServerConfigProvider) GetMultiUploadPartSize() (int64, bool) {
	for _, provider := range c.chain {
		val, ok := provider.GetMultiUploadPartSize()
		if ok {
			return val, true
		}
	}
	panic("There is no MultiUploadPartSize found!")
	return 0, false
}
