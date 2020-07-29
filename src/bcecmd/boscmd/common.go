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

package boscmd

import (
	"fmt"
)

import (
	"bceconf"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/util/log"
)

const (
	BOS_DEFAULT_DOMAIN = "bcebos.com"
	BOS_PATH_SEPARATOR = "/"
)

// Get endpoint from cache
func GetEndpointOfBucketFromCache(bucketName string) (string, bool) {
	log.Infof("Start to get endpoint of bucket %s from cache", bucketName)
	if endpoint, ok := bceconf.BucketEndpointCacheProvider.Get(bucketName); ok {
		log.Infof("Success get endpoint of bucket %s from cache, endpoint is %s", bucketName,
			endpoint)
		return endpoint, true
	}
	log.Infof("Failed to get endpoint of bucket %s from cache", bucketName)
	return "", false
}

// Get endpoint of bucket from bos
func GetEndpointOfBucketFromeBos(cli *bos.Client, bucketName string) (string, error) {
	log.Infof("Start to get endpoint of bucket %s from bos", bucketName)
	region, err := cli.GetBucketLocation(bucketName)
	if err != nil {
		log.Infof("Failed to get endpoint of bucket %s from bos, Error: %s", bucketName, err)
		return "", err
	}
	log.Infof("Success get region of bucket %s from bos, region is '%s'", bucketName, region)
	if region == "" {
		return "", fmt.Errorf("get a empty region from bos server!")
	}
	endpoint, _ := bceconf.ServerConfigProvider.GetDomainByRegion(region)
	bceconf.BucketEndpointCacheProvider.Write(bucketName, endpoint, 3600)
	return endpoint, nil
}

func GetEndpointOfBucket(cli *bos.Client, bucketName string) (string, error) {
	// get endpoint from cache
	endpoint, ok := GetEndpointOfBucketFromCache(bucketName)
	if ok {
		return endpoint, nil
	}

	// get endpoint from bos
	endpoint, err := GetEndpointOfBucketFromeBos(cli, bucketName)
	if err != nil {
		return "", err
	}
	return endpoint, nil
}
