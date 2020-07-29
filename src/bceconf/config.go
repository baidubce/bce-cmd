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

// This file provide the major operations on BCE configuration

package bceconf

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"utils/util"
)

const (
	BCE_VERSION                 = "0.3.0"
	EMPTY_STRING                = "none"
	DEFAULT_FOLDER_IN_USER_HOME = ".go-bcecli"
)

var (
	configFolder                string
	credentialPath              string
	configPath                  string
	bucktEndpointCachePath      string
	MultiuploadFolder           string
	credentialFileProvider      *FileCredentialProvider
	defaCredentialProvider      *DefaultCredentialProvider
	CredentialProvider          *ChainCredentialProvider
	serverConfigFileProvider    *FileServerConfigProvider
	defaServerConfigProvider    *DefaultServerConfigProvider
	ServerConfigProvider        *ChainServerConfigProvider
	BucketEndpointCacheProvider *BucketToEndpointCacheProvider
)

// If config folder don't exist, make config folder
func initConfFolder(configDirPath string) error {
	if ok := util.DoesDirExist(configDirPath); ok {
		return nil
	}
	return util.TryMkdir(configDirPath)
}

func InitConfig(configDirPath string) error {
	var (
		err error
	)
	// get the directory of config file
	if configDirPath == "" {
		configDirPath, err = util.GetHomeDirOfUser()
		if err != nil {
			return err
		}
		configDirPath = filepath.Join(configDirPath, DEFAULT_FOLDER_IN_USER_HOME)
	} else {
		configDirPath, err = util.Abs(configDirPath)
		if err != nil {
			return err
		}
	}

	// Check whether config have been initialized by read the same configuration file.
	if configFolder == configDirPath {
		return nil // have initialized
	} else {
		configFolder = configDirPath
	}
	err = initConfFolder(configDirPath)
	if err != nil {
		return err
	}

	credentialPath = filepath.Join(configDirPath, "credentials")
	configPath = filepath.Join(configDirPath, "config")
	bucktEndpointCachePath = filepath.Join(configDirPath, "bucket_endpoint_cache")
	MultiuploadFolder = filepath.Join(configDirPath, "multiupload_infos", "ak", "")

	// generate credential provider
	credentialFileProvider, err = NewFileCredentialProvider(credentialPath)
	if err != nil {
		return err
	}
	defaCredentialProvider, err = NewDefaultCredentialProvider()
	if err != nil {
		return err
	}
	CredentialProvider = NewChainCredentialProvider([]CredentialProviderInterface{
		credentialFileProvider, defaCredentialProvider})

	// generate server config provider
	serverConfigFileProvider, err = NewFileServerConfigProvider(configPath)
	if err != nil {
		return err
	}
	defaServerConfigProvider, err = NewDefaultServerConfigProvider()
	if err != nil {
		return err
	}
	ServerConfigProvider = NewChainServerConfigProvider([]ServerConfigProviderInterface{
		serverConfigFileProvider, defaServerConfigProvider})

	// genrate cahce provider
	BucketEndpointCacheProvider, err = NewBucketToEndpointCacheProvider(bucktEndpointCachePath)
	if err != nil {
		return err
	}
	return nil

}

// Provide interactive configuration with user
func ConfigInteractive(configDirPath string) {
	var (
		newAk                       string
		newSk                       string
		newSts                      string
		newRegion                   string
		newDomain                   string
		newUseAutoSwitchDomain      string
		newBreakpointFileExpiration string
		newUseHttps                 string
		newMultiUploadThreadNum     string
		newSyncProcessingNum        string
		newMultiUploadPartSize      string
	)

	// Init Configuration info
	if err := InitConfig(configDirPath); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	// Config ak
	ak, ok := CredentialProvider.GetAccessKey()
	if !ok || ak == "" {
		ak = EMPTY_STRING
	}
	fmt.Printf("Your Access Key ID [%s]: ", ak)
	scanner.Scan()
	newAk = strings.TrimSpace(scanner.Text())
	if newAk != "" {
		if strings.ToLower(newAk) == EMPTY_STRING {
			newAk = ""
		}
		credentialFileProvider.SetAccessKey(newAk)
	}

	// Config sk
	sk, ok := CredentialProvider.GetSecretKey()
	if !ok || sk == "" {
		sk = EMPTY_STRING
	}
	fmt.Printf("Your Secret Access Key ID [%s]: ", sk)
	scanner.Scan()
	newSk = strings.TrimSpace(scanner.Text())
	if newSk != "" {
		if strings.ToLower(newSk) == EMPTY_STRING {
			newSk = ""
		}
		credentialFileProvider.SetSecretKey(newSk)
	}

	stsToken, ok := CredentialProvider.GetSecurityToken()
	fmt.Printf("Your Security Token [%s]: ", stsToken)
	scanner.Scan()
	newSts = strings.TrimSpace(scanner.Text())
	credentialFileProvider.SetSecurityToken(newSts)

	// Config region
	region, ok := ServerConfigProvider.GetRegion()
	if !ok || region == "" {
		region = EMPTY_STRING
	}
	fmt.Printf("Default Region Name [%s]: ", region)
	scanner.Scan()
	newRegion = strings.TrimSpace(scanner.Text())
	if newRegion != "" {
		if strings.ToLower(newRegion) == EMPTY_STRING {
			newRegion = ""
		}
		serverConfigFileProvider.SetRegion(newRegion)
	}

	// Config domain
	domain, ok := ServerConfigProvider.GetDomain()
	if !ok || domain == "" {
		domain = EMPTY_STRING
	}
	fmt.Printf("Default Domain [%s]: ", domain)
	scanner.Scan()
	newDomain = strings.TrimSpace(scanner.Text())
	if newDomain != "" {
		if strings.ToLower(newDomain) == EMPTY_STRING {
			newDomain = ""
		}
		serverConfigFileProvider.SetDomain(newDomain)
	}

	// Config use auto switch domain
	var propmtUseAutoSwitchDomain string
	useAutoSwitchDomain, ok := ServerConfigProvider.GetUseAutoSwitchDomain()
	if !ok {
		propmtUseAutoSwitchDomain = EMPTY_STRING
	} else {
		propmtUseAutoSwitchDomain = BOOL_TO_STRING[useAutoSwitchDomain]
	}
	fmt.Printf("Default use auto switch domain [%s]", propmtUseAutoSwitchDomain)
	scanner.Scan()
	newUseAutoSwitchDomain = strings.TrimSpace(scanner.Text())
	if newUseAutoSwitchDomain != "" {
		newUseAutoSwitchDomain = strings.ToLower(newUseAutoSwitchDomain)
		if newUseAutoSwitchDomain == EMPTY_STRING {
			newUseAutoSwitchDomain = ""
		} else {
			if _, valOk := AOLLOWED_CONFIRM_OPTIONS[newUseAutoSwitchDomain]; !valOk {
				fmt.Printf("Only support 'no' and 'yes', [%s] is invalid, default value is used.\n",
					newUseAutoSwitchDomain)
				newUseAutoSwitchDomain = ""
			}
		}
		serverConfigFileProvider.SetUseAutoSwitchDomain(newUseAutoSwitchDomain)
	}

	// Config breakpoint file expireation
	var propmtBreakpointExpir string
	if breakpointExpir, ok := ServerConfigProvider.GetBreakpointFileExpiration(); ok {
		propmtBreakpointExpir = strconv.Itoa(breakpointExpir)
	} else {
		propmtBreakpointExpir = EMPTY_STRING
	}
	fmt.Printf("Default breakpoint_file_expiration [%s] days: ", propmtBreakpointExpir)
	scanner.Scan()
	newBreakpointFileExpiration = strings.TrimSpace(scanner.Text())
	if newBreakpointFileExpiration != "" {
		if strings.ToLower(newBreakpointFileExpiration) == EMPTY_STRING {
			newBreakpointFileExpiration = ""
		} else {
			if _, ok := strconv.Atoi(newBreakpointFileExpiration); ok != nil {
				fmt.Printf("File expiration must be positive integer, [%s] is not valid."+
					" default value is used.\n", newBreakpointFileExpiration)
				newBreakpointFileExpiration = ""
			}
		}
		serverConfigFileProvider.SetBreakpointFileExpiration(newBreakpointFileExpiration)
	}

	// Config use https protocol
	var propmtUseHttps string
	useHttps, ok := ServerConfigProvider.GetUseHttpsProtocol()
	if !ok {
		propmtUseHttps = EMPTY_STRING
	} else {
		propmtUseHttps = BOOL_TO_STRING[useHttps]
	}
	fmt.Printf("Default use https protocol [%s]", propmtUseHttps)
	scanner.Scan()
	newUseHttps = strings.TrimSpace(scanner.Text())
	if newUseHttps != "" {
		newUseHttps = strings.ToLower(newUseHttps)
		if newUseHttps == EMPTY_STRING {
			newUseHttps = ""
		} else {
			if _, ok := AOLLOWED_CONFIRM_OPTIONS[newUseHttps]; !ok {
				fmt.Printf("Only support 'no' and 'yes', [%s] is invalid, default value is used.\n",
					newUseHttps)
				newUseHttps = ""
			}
		}
		serverConfigFileProvider.SetUseHttpsProtocol(newUseHttps)
	}

	// Config multi upload thread num
	var propmtMultiUploadThreadNum string
	if multiUploadThreadNum, ok := ServerConfigProvider.GetMultiUploadThreadNum(); ok {
		propmtMultiUploadThreadNum = strconv.FormatInt(multiUploadThreadNum, 10)
	} else {
		propmtMultiUploadThreadNum = EMPTY_STRING
	}
	fmt.Printf("Default multi upload thread num [%s]: ", propmtMultiUploadThreadNum)
	scanner.Scan()
	newMultiUploadThreadNum = strings.TrimSpace(scanner.Text())
	if newMultiUploadThreadNum != "" {
		if strings.ToLower(newMultiUploadThreadNum) == EMPTY_STRING {
			newMultiUploadThreadNum = ""
		} else {
			if _, ok := strconv.Atoi(newMultiUploadThreadNum); ok != nil {
				fmt.Printf("Input multi upload thread num must be a positive integer, [%s] is"+
					" not valid, default multi upload thread num [%s] is used\n",
					newMultiUploadThreadNum, propmtMultiUploadThreadNum)
				newMultiUploadThreadNum = ""
			}
		}
		serverConfigFileProvider.SetMultiUploadThreadNum(newMultiUploadThreadNum)
	}

	// Config sync processing num
	var propmtSyncProcessingNum string
	if syncProcessingNum, ok := ServerConfigProvider.GetSyncProcessingNum(); ok {
		propmtSyncProcessingNum = strconv.Itoa(syncProcessingNum)
	} else {
		propmtSyncProcessingNum = EMPTY_STRING
	}
	fmt.Printf("Default sync processing num [%s]: ", propmtSyncProcessingNum)
	scanner.Scan()
	newSyncProcessingNum = strings.TrimSpace(scanner.Text())
	if newSyncProcessingNum != "" {
		if strings.ToLower(newSyncProcessingNum) == EMPTY_STRING {
			newSyncProcessingNum = ""
		} else {
			if _, ok := strconv.Atoi(newSyncProcessingNum); ok != nil {
				fmt.Printf("Input sync processing num must be a positive integer, [%s] is"+
					" not valid, default value [%s] is used\n",
					newSyncProcessingNum, propmtSyncProcessingNum)
				newSyncProcessingNum = ""
			}
		}
		serverConfigFileProvider.SetSyncProcessingNum(newSyncProcessingNum)
	}

	// Config multi upload part size
	var propmtMultiUploadPartSize string
	if multiUploadPartSize, ok := ServerConfigProvider.GetMultiUploadPartSize(); ok {
		propmtMultiUploadPartSize = strconv.FormatInt(multiUploadPartSize, 10)
	} else {
		propmtMultiUploadPartSize = EMPTY_STRING
	}
	fmt.Printf("Default multi upload part size [%s] MB (Must be positive integer and equal or "+
		"greater than 1) : ", propmtMultiUploadPartSize)
	scanner.Scan()
	newMultiUploadPartSize = strings.TrimSpace(scanner.Text())
	if newMultiUploadPartSize != "" {
		if strings.ToLower(newMultiUploadPartSize) == EMPTY_STRING {
			newMultiUploadPartSize = ""
		} else {
			if val, ok := strconv.ParseInt(newMultiUploadPartSize, 10, 64); ok != nil || val < 1 {
				fmt.Printf("Input multi upload part size must be a positive integer and equal or "+
					"greater than 1, [%s] is not valid, default multi upload part size [%s] "+
					"is used\n", newMultiUploadPartSize, propmtMultiUploadPartSize)
				newMultiUploadPartSize = ""
			}
		}
		serverConfigFileProvider.SetMultiUploadPartSize(newMultiUploadPartSize)
	}

	credentialFileProvider.save()
	serverConfigFileProvider.save()
}

// users can use --conf-path to reload configuration from specified path.
func ReloadConfAction(configDirPath string) {
	err := InitConfig(configDirPath)
	if err != nil {
		panic(err)
	}
}

// If the conf of cache and config are changed.
// We do not save the change to file, when run time.
// We save the change into file when exit!
func DestructConfFolder() {
	if serverConfigFileProvider != nil {
		serverConfigFileProvider.save()
	}
	if BucketEndpointCacheProvider != nil {
		BucketEndpointCacheProvider.save()
	}
	if credentialFileProvider != nil {
		credentialFileProvider.save()
	}
}
