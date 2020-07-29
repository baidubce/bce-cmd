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

// This file provide cache for bucket_name => endpoint

package bceconf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

import (
	"utils/util"
)

const (
	BUCKET_CACHE_EXPIRE = 3600
)

type BucketToEndpointInfo struct {
	Endpoint string
}

type BucketToEndpointCache struct {
	Buckets map[string]*BucketToEndpointInfo
}

// New bucket to endpoint cahce  provider
func NewBucketToEndpointCacheProvider(cachePath string) (*BucketToEndpointCacheProvider, error) {
	b := &BucketToEndpointCacheProvider{
		storeCachePath: cachePath,
	}
	err := b.loadingCachce()
	if err == nil {
		return b, nil
	}
	return nil, err
}

type BucketToEndpointCacheProvider struct {
	storeCachePath string
	dirty          bool
	cache          *BucketToEndpointCache
	rwmutex        sync.RWMutex
}

func (b *BucketToEndpointCacheProvider) loadingCachce() error {
	b.cache = &BucketToEndpointCache{}
	b.rwmutex.RLock()
	defer b.rwmutex.RUnlock()
	if ok := util.DoesFileExist(b.storeCachePath); ok {
		if err := LoadConfig(b.storeCachePath, b.cache); err == nil {
			return nil
		}
		// we have failed to load this cache file, maybe it is disrupted
		// we just delete it
		if err := os.Remove(b.storeCachePath); err != nil {
			return fmt.Errorf("failed to delete the disrupted bucket-to-enpoint cache file! %v",
				err)
		}
	}
	return nil
}

// the format of cache is domain|timestamp, such as "bj.bcebos.com|456453343"
func (b *BucketToEndpointCacheProvider) splitRegionDomainAndExpire(val string) (string, int64,
	bool) {
	if val == "" {
		return "", 0, false
	}
	val = strings.TrimSpace(val)
	values := strings.Split(val, "|")
	if len(values) != 2 {
		return "", 0, false
	} else if values[0] == "" {
		return "", 0, false
	}

	expireTime, ok := strconv.ParseInt(values[1], 10, 64)
	if ok != nil || expireTime < 0 {
		return "", 0, false
	}
	return values[0], expireTime, true
}

func (b *BucketToEndpointCacheProvider) timeIsExpired(expireTime int64) bool {
	now := time.Now().Unix()
	if expireTime < now {
		return true
	}
	return false
}

func (b *BucketToEndpointCacheProvider) save() error {
	if b.storeCachePath == "" {
		return fmt.Errorf("The path of cache file is emtpy")
	}
	if !b.dirty {
		return nil
	}
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	err := WriteConfig(b.storeCachePath, b.cache)
	if err == nil {
		b.dirty = false
	}
	return err
}

// Get the endpoint of bucket from cache.
// when endpoint is expired, delete the bucket from cache
func (b *BucketToEndpointCacheProvider) Get(bucketName string) (string, bool) {
	if bucketName == "" {
		return "", false
	}

	b.rwmutex.RLock()
	val, ok := b.cache.Buckets[bucketName]
	b.rwmutex.RUnlock()
	if !ok {
		return "", false
	}
	if val != nil {
		endpoint, expireTime, splintOk := b.splitRegionDomainAndExpire(val.Endpoint)
		if splintOk {
			if !b.timeIsExpired(expireTime) {
				return endpoint, true
			}
		}
	}
	// delete info of bucket from cache when val is nil or expired.
	b.Delete(bucketName)
	return "", false
}

// Write the endpoint of bucket to cache
func (b *BucketToEndpointCacheProvider) Write(bucketName, endpoint string, expireTime int64) bool {
	if bucketName == "" || endpoint == "" {
		return false
	}
	if expireTime <= 0 {
		expireTime = BUCKET_CACHE_EXPIRE
	}
	timestampOfNow := time.Now().Unix()
	values := endpoint + "|" + strconv.FormatInt(timestampOfNow+expireTime, 10)
	bucketEndpointVal := &BucketToEndpointInfo{Endpoint: values}
	b.rwmutex.Lock()
	defer b.rwmutex.Unlock()
	if b.cache.Buckets == nil {
		b.cache.Buckets = make(map[string]*BucketToEndpointInfo)
	}
	b.cache.Buckets[bucketName] = bucketEndpointVal
	b.dirty = true
	return true
}

// Delete the endpoint of bucket from cache
func (b *BucketToEndpointCacheProvider) Delete(bucketName string) error {
	if bucketName != "" {
		b.rwmutex.Lock()
		defer b.rwmutex.Unlock()
		delete(b.cache.Buckets, bucketName)
		b.dirty = true
	}
	return nil
}
