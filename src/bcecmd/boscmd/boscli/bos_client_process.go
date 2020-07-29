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

// Codes in this file is used to process bosClient.

package boscli

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/auth"
	"github.com/baidubce/bce-sdk-go/util/log"
	"runtime"
	"strings"
	"sync"
)

import (
	"bcecmd/boscmd"
	"bceconf"
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
)

const (
	BOSCLIENT_RETRY_NUM       = 3
	BOSCLIENT_RETRY_MAX_DELAY = 20000
	BOSCLIENT_RETRY_BASE      = 300
)

var (
	defaultBosClient *bos.Client
	endpointRWMutex  sync.RWMutex
)

func getUserAgent() string {
	var userAgent string
	userAgent += BCE_CLI_AGENT
	userAgent += "/" + bceconf.BCE_VERSION
	userAgent += "/" + runtime.Version()
	userAgent += "/" + runtime.GOOS
	userAgent += "/" + runtime.GOARCH
	return userAgent
}

func setEndpointProtocol(endpoint string, useHttps bool) (string, error) {
	if strings.HasPrefix(endpoint, HTTP_PROTOCOL) {
		endpoint = endpoint[len(HTTP_PROTOCOL):]
	} else if strings.HasPrefix(endpoint, HTTPS_PROTOCOL) {
		endpoint = endpoint[len(HTTPS_PROTOCOL):]
	}
	if endpoint == "" {
		return "", fmt.Errorf("Endpoint is empty!")
	}
	if useHttps {
		return HTTPS_PROTOCOL + endpoint, nil
	}
	return HTTP_PROTOCOL + endpoint, nil
}

func newBosClient(ak, sk, endpoint string, useHttps bool, multiPartSize,
	multiParallerl int64) (*bos.Client, error) {

	// set http or https protocol
	endpoint, err := setEndpointProtocol(endpoint, useHttps)
	if err != nil {
		return nil, fmt.Errorf("Endpoint is invalid!")
	}
	bosClient, err := bos.NewClient(ak, sk, endpoint)
	if err != nil {
		return nil, err
	}
	bosClient.Config.Retry = bce.NewBackOffRetryPolicy(BOSCLIENT_RETRY_NUM,
		BOSCLIENT_RETRY_MAX_DELAY, BOSCLIENT_RETRY_BASE)
	bosClient.Config.UserAgent = getUserAgent()

	// set parallel num and part size for uploading super file.
	bosClient.MaxParallel = multiParallerl
	bosClient.MultipartSize = multiPartSize

	return bosClient, nil
}

// Init bos client.
func buildBosClient(ak, sk, endpoint string,
	credentialProvider bceconf.CredentialProviderInterface,
	serverConfigProvider bceconf.ServerConfigProviderInterface) (*bos.Client, error) {
	var (
		ok bool
	)

	if ak == "" || sk == "" {
		if ak, ok = credentialProvider.GetAccessKey(); !ok {
			return nil, fmt.Errorf("There is no access key found!")
		}
		if sk, ok = credentialProvider.GetSecretKey(); !ok {
			return nil, fmt.Errorf("There is no access secret key found!")
		}
	}
	if endpoint == "" {
		if endpoint, ok = serverConfigProvider.GetDomain(); !ok {
			return nil, fmt.Errorf("There is no endpoint found!")
		}
	}

	stsToken, _ := credentialProvider.GetSecurityToken()

	// get multi upload part size
	multiUploadPartSize, ok := serverConfigProvider.GetMultiUploadPartSize()
	if !ok {
		return nil, fmt.Errorf("There is no info about multi upload part size found!")
	}

	multiUploadPartSizeByte := multiUploadPartSize * (1 << 20)

	// get multi upload or copy thread num
	multiUploadThreadNum, ok := serverConfigProvider.GetMultiUploadThreadNum()
	if !ok {
		return nil, fmt.Errorf("There is no info about multi upload thread Num found!")
	}
	if useHttps, ok := serverConfigProvider.GetUseHttpsProtocol(); !ok {
		return nil, fmt.Errorf("There is no https protocol info found!")
	} else {

		bosClient, err := newBosClient(ak, sk, endpoint, useHttps, multiUploadPartSizeByte, multiUploadThreadNum)
		if err != nil {
			return bosClient, nil
		}
		if stsToken == bceconf.DEFAULT_STS {
			return bosClient, nil
		}
		stsCredential, err := auth.NewSessionBceCredentials(
			ak,
			sk,
			stsToken)
		if err != nil {
			fmt.Println("create sts credential object failed:", err)
			return nil, err
		}
		bosClient.Config.Credentials = stsCredential
		return bosClient, nil
	}
}

// init BosCLient
func bosClientInit(ak, sk, endpoint string) (bosClientInterface, error) {
	bosClient, err := buildBosClient(
		ak,
		sk,
		endpoint,
		bceconf.CredentialProvider,
		bceconf.ServerConfigProvider,
	)
	if err != nil {
		return nil, err
	}
	if useAutoSD, ok := bceconf.ServerConfigProvider.GetUseAutoSwitchDomain(); ok && useAutoSD {
		return &bosClientWrapper{bosClient: bosClient}, nil
	}
	return bosClient, nil
}

// build a new bos client, endpoint is the endpoint of bucketName
func initBosClientForBucket(ak, sk, bucketName string) (bosClientInterface, error) {
	bosClient, err := buildBosClient(
		ak,
		sk,
		"",
		bceconf.CredentialProvider,
		bceconf.ServerConfigProvider,
	)
	if err != nil {
		return nil, err
	}

	// when auto switch domian is enabled, change endpoint
	if useAutoSD, ok := bceconf.ServerConfigProvider.GetUseAutoSwitchDomain(); ok && useAutoSD {
		if err = modifiyBosClientEndpointByBucketName(bosClient, bucketName); err != nil {
			return nil, err
		}
		return &bosClientWrapper{bosClient: bosClient}, nil
	}
	return bosClient, nil
}

// midifiy endpoint of bos client by endpoint
func modifiyBosClientEndpointByEndpoint(bosClient *bos.Client, endpoint string) error {
	log.Debugf("get user https protocol")
	useHttps, ok := bceconf.ServerConfigProvider.GetUseHttpsProtocol()

	if !ok {
		return fmt.Errorf("There is no https protocol info found!")
	}
	log.Debugf("set end point protocol")
	endpoint, err := setEndpointProtocol(endpoint, useHttps)
	if err != nil {
		return fmt.Errorf("Endpoint is invalid!")
	}
	log.Debugf("endpoint lock")
	endpointRWMutex.RLock()
	if bosClient.Config.Endpoint != endpoint {
		endpointRWMutex.RUnlock()
		endpointRWMutex.Lock()
		bosClient.Config.Endpoint = endpoint
		endpointRWMutex.Unlock()
	} else {
		endpointRWMutex.RUnlock()
	}
	log.Debugf("endpoint unlock")
	return nil
}

// midifiy endpoint of bos client by bucket name, gen endpoint from cache or bos
func modifiyBosClientEndpointByBucketName(bosClient *bos.Client, bucketName string) error {
	endpoint, err := boscmd.GetEndpointOfBucket(bosClient, bucketName)
	if err != nil {
		return err
	}
	return modifiyBosClientEndpointByEndpoint(bosClient, endpoint)
}

// midifiy endpoint of bos client by bucket name, get endpoint from cache
func modifiyBosClientEndpointByBucketNameCache(bosClient *bos.Client, bucketName string) bool {
	if endpoint, ok := boscmd.GetEndpointOfBucketFromCache(bucketName); ok {
		err := modifiyBosClientEndpointByEndpoint(bosClient, endpoint)
		if err == nil {
			return true
		}
	}
	return false
}

// midifiy endpoint of bos client by bucket name, get endpoint from bos
func modifiyBosClientEndpointByBucketNameBos(bosClient *bos.Client, bucketName string) error {
	endpoint, err := boscmd.GetEndpointOfBucketFromeBos(bosClient, bucketName)
	if err != nil {
		return err
	}
	return modifiyBosClientEndpointByEndpoint(bosClient, endpoint)
}
