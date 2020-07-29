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
	"strconv"
	"testing"
	"time"
)

import (
	"utils/util"
)

var (
	bCacehProvider1 = &BucketToEndpointCacheProvider{
		cache: &BucketToEndpointCache{
			Buckets: map[string]*BucketToEndpointInfo{
				"bj": &BucketToEndpointInfo{
					Endpoint: "bj.bcebos.com|1512655719",
				},
				"gz": &BucketToEndpointInfo{
					Endpoint: "gz.bcebos.com|1512655719",
				},
			},
		},
	}
	bCacehProvider2 = &BucketToEndpointCacheProvider{
		cache: &BucketToEndpointCache{},
	}
	rightTime       = strconv.FormatInt(time.Now().Unix()+1000, 10)
	expiredTime     = strconv.FormatInt(time.Now().Unix()-1000, 10)
	nowTime         = strconv.FormatInt(time.Now().Unix(), 10)
	bCacehProvider3 = &BucketToEndpointCacheProvider{
		cache: &BucketToEndpointCache{
			Buckets: map[string]*BucketToEndpointInfo{
				"right": &BucketToEndpointInfo{
					Endpoint: "right.bcebos.com|" + rightTime,
				},
				"expired": &BucketToEndpointInfo{
					Endpoint: "expired.bcebos.com|" + expiredTime,
				},
				"now": &BucketToEndpointInfo{
					Endpoint: "now.bcebos.com|" + nowTime,
				},
			},
		},
	}
)

func init() {
	err := util.TryMkdir("./test_file/")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file")
		os.Exit(2)
	}
	fd, err := os.Create("./test_file/test_cache1.cache")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file/test_cache1.cache")
		os.Exit(2)
	}
	cacheVal := "[Buckets \"liupeng-bj\"]\nEndpoint = bj.bcebos.com|1512655719\n" +
		"[Buckets \"liupeng-gz\"]\nEndpoint = gz.bcebos.com|1512655719\n"
	fmt.Fprintf(fd, cacheVal)
	fd.Close()

	fd, err = os.Create("./test_file/test_cache_wrong.cache")
	if err != nil {
		fmt.Printf("Fail to create dir ./test_file/test_cache_wrong.cache")
		os.Exit(2)
	}
	cacheVal = "[Bucket \"liupeng-bj\"]\nEndpoint = bj.bcebos.com|1512655719\n" +
		"[Buckets \"liupeng-gz\"]\nEndpoint = gz.bcebos.com|1512655719\n"
	fmt.Fprintf(fd, cacheVal)
	fd.Close()
}

type newBucketToEndpointCacheProviderType struct {
	path  string
	isSuc bool
}

func TestNewBucketToEndpointCacheProvider(t *testing.T) {
	testCases := []newBucketToEndpointCacheProviderType{
		newBucketToEndpointCacheProviderType{
			path:  "./xxxx.cfg",
			isSuc: true,
		},
		newBucketToEndpointCacheProviderType{
			path:  "./test_file/test_cache1.cache",
			isSuc: true,
		},
		newBucketToEndpointCacheProviderType{
			path:  "./test_file/test_cache_wrong.cache",
			isSuc: false,
		},
	}
	for i, tCase := range testCases {
		ret, err := NewBucketToEndpointCacheProvider(tCase.path)
		util.ExpectEqual("cache NewBucketToEndpointCacheProvider I", i+1, t.Errorf, tCase.isSuc,
			err == nil)
		if tCase.isSuc == true && err != nil {
			t.Logf("id %d error: %s\n", i+1, err)
		}
		if err == nil && ret == nil {
			t.Errorf("cache NewBucketToEndpointCacheProvider II id %d, err is nil but ret ==nil",
				i+1)
		}
	}
}

type bucketCaceLoadingType struct {
	provider  *BucketToEndpointCacheProvider
	oProvider *BucketToEndpointCacheProvider
	isSuc     bool
}

func TestloadingCachce(t *testing.T) {

	testCases := []bucketCaceLoadingType{
		bucketCaceLoadingType{
			provider: &BucketToEndpointCacheProvider{
				storeCachePath: "./test_file/test_cache1.cache",
			},
			oProvider: bCacehProvider1,
			isSuc:     true,
		},
		bucketCaceLoadingType{
			provider: &BucketToEndpointCacheProvider{
				storeCachePath: "./test_file/test_cache_wrong.cache",
			},
			isSuc: false,
		},
		bucketCaceLoadingType{
			provider: &BucketToEndpointCacheProvider{
				storeCachePath: "./test_file/test_cache_xxx.cache",
			},
			oProvider: bCacehProvider2,
			isSuc:     true,
		},
	}

	for i, tCase := range testCases {
		err := tCase.provider.loadingCachce()
		util.ExpectEqual("cache loadingCachce I", i+1, t.Errorf, tCase.isSuc,
			err != nil)
		if tCase.isSuc {
			if len(tCase.oProvider.cache.Buckets) == 0 {
				util.ExpectEqual("cache loadingCachce II", i+1, t.Errorf, 0,
					len(tCase.provider.cache.Buckets))
			} else {
				util.ExpectEqual("cache loadingCachce III", i+1, t.Errorf,
					tCase.oProvider.cache.Buckets, tCase.provider.cache.Buckets)
			}
		}
	}
}

type splitRegionDomainAndExpireType struct {
	input  string
	domain string
	expire int64
	isSuc  bool
}

func TestSplitRegionDomainAndExpire(t *testing.T) {
	testCases := []splitRegionDomainAndExpireType{
		splitRegionDomainAndExpireType{
			input: "bj.bcebos.com|",
			isSuc: false,
		},
		splitRegionDomainAndExpireType{
			input:  "bj.bcebos.com|123",
			domain: "bj.bcebos.com",
			expire: 123,
			isSuc:  true,
		},
		splitRegionDomainAndExpireType{
			input:  " bj.bcebos.com|123  ",
			domain: "bj.bcebos.com",
			expire: 123,
			isSuc:  true,
		},
		splitRegionDomainAndExpireType{
			input: "|123",
			isSuc: false,
		},
		splitRegionDomainAndExpireType{
			input: "|",
			isSuc: false,
		},
		splitRegionDomainAndExpireType{
			input: "",
			isSuc: false,
		},
		splitRegionDomainAndExpireType{
			input: "   |   ",
			isSuc: false,
		},
	}
	for i, tCase := range testCases {
		domain, expire, ok := bCacehProvider2.splitRegionDomainAndExpire(tCase.input)
		util.ExpectEqual("cache splitRegionDomainAndExpire I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("cache splitRegionDomainAndExpire II", i+1, t.Errorf, tCase.domain,
				domain)
			util.ExpectEqual("cache splitRegionDomainAndExpire III", i+1, t.Errorf, tCase.expire,
				expire)
		}
	}
}

func TestTimeIsExpired(t *testing.T) {
	ok := bCacehProvider2.timeIsExpired(0)
	util.ExpectEqual("cache timeIsExpired I", 1, t.Errorf, true, ok)
	ok = bCacehProvider2.timeIsExpired(time.Now().Unix() + 100)
	util.ExpectEqual("cache timeIsExpired I", 2, t.Errorf, false, ok)
}

type bCacheSaveType struct {
	provider *BucketToEndpointCacheProvider
	path     string
	setDirty bool
	isErr    bool
	dirty    bool
}

func TestBCacheSave(t *testing.T) {
	bCacehProvider1.dirty = false
	bCacehProvider2.dirty = false
	testCases := []bCacheSaveType{
		bCacheSaveType{
			provider: bCacehProvider1,
			path:     "",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		bCacheSaveType{
			provider: bCacehProvider1,
			path:     "/root/cfg",
			setDirty: true,
			isErr:    true,
			dirty:    true,
		},
		bCacheSaveType{
			provider: bCacehProvider1,
			path:     "./test.cfg",
			setDirty: true,
			isErr:    false,
			dirty:    false,
		},
		bCacheSaveType{
			provider: bCacehProvider2,
			path:     "./test.cfg",
			setDirty: true,
			isErr:    false,
			dirty:    false,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = tCase.setDirty
		tCase.provider.storeCachePath = tCase.path
		err := tCase.provider.save()
		util.ExpectEqual("cache save I", i+1, t.Errorf, tCase.isErr, err != nil)
		if tCase.isErr == false && err != nil {
			t.Logf("error: %s", err)
		}
		util.ExpectEqual("cache save II", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
	}
}

type getCacheType struct {
	provider *BucketToEndpointCacheProvider
	bucket   string
	endpoint string
	isSuc    bool
}

func TestGetEndpointOfBucket(t *testing.T) {
	testCases := []getCacheType{
		getCacheType{
			provider: bCacehProvider3,
			bucket:   "right",
			endpoint: "right.bcebos.com",
			isSuc:    true,
		},
		getCacheType{
			provider: bCacehProvider3,
			bucket:   "expired",
			isSuc:    false,
		},
		getCacheType{
			provider: bCacehProvider3,
			bucket:   "dontexist",
			isSuc:    false,
		},
		getCacheType{
			provider: bCacehProvider3,
			bucket:   "now",
			endpoint: "now.bcebos.com",
			isSuc:    true,
		},
		getCacheType{
			provider: bCacehProvider3,
			bucket:   "",
			isSuc:    false,
		},
	}
	for i, tCase := range testCases {
		endpoint, ok := tCase.provider.Get(tCase.bucket)
		util.ExpectEqual("cache Get I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			util.ExpectEqual("cache Get I", i+1, t.Errorf, tCase.endpoint, endpoint)
		}
	}
}

type writeCacheType struct {
	provider   *BucketToEndpointCacheProvider
	bucket     string
	endpoint   string
	expireTime int64
	isSuc      bool
}

func TestWriteEndpointOfBucket(t *testing.T) {
	testCases := []writeCacheType{
		writeCacheType{
			provider:   bCacehProvider1,
			bucket:     "bj",
			endpoint:   "xbj.bcebos.com",
			expireTime: 1800,
			isSuc:      true,
		},
		writeCacheType{
			provider: bCacehProvider1,
			bucket:   "gz",
			endpoint: "xgz.bcebos.com",
			isSuc:    true,
		},
		writeCacheType{
			provider:   bCacehProvider1,
			bucket:     "su",
			endpoint:   "xsu.bcebos.com",
			expireTime: -100,
			isSuc:      true,
		},
		writeCacheType{
			provider:   bCacehProvider1,
			bucket:     "now",
			expireTime: 1800,
			isSuc:      false,
		},
		writeCacheType{
			provider:   bCacehProvider1,
			expireTime: 1800,
			isSuc:      false,
		},
		writeCacheType{
			provider: bCacehProvider1,
			isSuc:    false,
		},
		writeCacheType{
			provider: &BucketToEndpointCacheProvider{
				cache: &BucketToEndpointCache{},
			},
			bucket:   "su",
			endpoint: "xsu.bcebos.com",
			isSuc:    true,
		},
	}
	for i, tCase := range testCases {
		ok := tCase.provider.Write(tCase.bucket, tCase.endpoint, tCase.expireTime)
		util.ExpectEqual("cache write I", i+1, t.Errorf, tCase.isSuc, ok)
		if tCase.isSuc {
			endpoint, ok := tCase.provider.Get(tCase.bucket)
			if !ok {
				t.Errorf("get endpoint of %s failed", tCase.bucket)
			} else {
				util.ExpectEqual("cache write II", i+1, t.Errorf, tCase.endpoint, endpoint)
			}
		}
	}
}

type deleteCacheType struct {
	provider *BucketToEndpointCacheProvider
	bucket   string
	dirty    bool
	isSuc    bool
}

func TestDeleteEndpointOfBucket(t *testing.T) {
	testCases := []deleteCacheType{
		deleteCacheType{
			provider: bCacehProvider1,
			bucket:   "bj",
			dirty:    true,
			isSuc:    true,
		},
		deleteCacheType{
			provider: bCacehProvider1,
			bucket:   "gz",
			dirty:    true,
			isSuc:    true,
		},
		deleteCacheType{
			provider: bCacehProvider1,
			bucket:   "su",
			dirty:    true,
			isSuc:    true,
		},
		deleteCacheType{
			provider: bCacehProvider1,
			bucket:   "xxxxxx",
			dirty:    true,
			isSuc:    true,
		},
		deleteCacheType{
			provider: bCacehProvider1,
			bucket:   "adfsaf",
			dirty:    true,
			isSuc:    true,
		},
		deleteCacheType{
			provider: bCacehProvider1,
			dirty:    false,
			isSuc:    true,
		},
	}
	for i, tCase := range testCases {
		tCase.provider.dirty = false
		err := tCase.provider.Delete(tCase.bucket)
		util.ExpectEqual("cache delete I", i+1, t.Errorf, tCase.isSuc, err == nil)
		util.ExpectEqual("cache delete II", i+1, t.Errorf, tCase.dirty, tCase.provider.dirty)
		if tCase.isSuc {
			_, ok := tCase.provider.Get(tCase.bucket)
			util.ExpectEqual("cache delete III", i+1, t.Errorf, false, ok)
		}
	}
}
