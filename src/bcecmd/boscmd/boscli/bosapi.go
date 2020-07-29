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

// This module provides the major operations on BOS API.

package boscli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	// 	"path/filepath"
	// 	"strings"
)

import (
	// 	"bceconf"
	"bcecmd/boscmd"
	"utils/util"
	// 	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

// Create new BosApi
func NewBosApi() *BosApi {
	var (
		ak       string
		sk       string
		endpoint string
		err      error
	)

	bosapiClient := &BosApi{}
	bosapiClient.bosClient, err = bosClientInit(ak, sk, endpoint)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
	return bosapiClient
}

type BosApi struct {
	bosClient bosClientInterface
}

type putBucketAclArgs struct {
	bucketName string
	acl        []byte
	opType     int
}

// Put ACL
func (b *BosApi) PutBucketAcl(aclConfigPath, bosPath string, canned string) {

	// preprocessing
	// opType:
	//    1 put acl from file
	//    2 put acl from canned
	args, err, retCode := b.putBucketAclPreProcess(aclConfigPath, bosPath, canned)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}

	//executing
	err, retCode = b.putBucketAclExecute(args.opType, args.acl, args.bucketName, canned)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}
}

// Put ACL preprocessing
func (b *BosApi) putBucketAclPreProcess(aclConfigPath, bosPath, canned string) (*putBucketAclArgs,
	error, BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return nil, nil, retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return nil, nil, BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return nil, nil, BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	if aclConfigPath != "" && canned != "" {
		return nil, fmt.Errorf("Can't put acl from canned and file at the same time"),
			BOSCLI_PUT_ACL_CANNED_FILE_SAME_TIME
	}

	// is canned acl?
	if canned != "" {
		if canned != "private" && canned != "public-read" && canned != "public-read-write" {
			return nil, fmt.Errorf("usupported canned ACL"), BOSCLI_PUT_ACL_CANNED_DONT_SUPPORT
		}
		return &putBucketAclArgs{bucketName: bucketName, opType: 2}, nil, BOSCLI_OK
	}

	// config path exist?
	if aclConfigPath != "" {
		if !util.DoesFileExist(aclConfigPath) {
			return nil, nil, boscmd.LOCAL_FILE_NOT_EXIST
		}
		// acl is valid?
		fd, err := os.Open(aclConfigPath)
		if err != nil {
			return nil, err, BOSCLI_EMPTY_CODE
		}
		defer fd.Close()

		ruleJosn, err := ioutil.ReadAll(fd)
		if err != nil {
			return nil, err, BOSCLI_EMPTY_CODE
		}

		rule := &api.PutBucketAclArgs{}
		acl := json.NewDecoder(bytes.NewReader(ruleJosn))
		if err := acl.Decode(rule); err != nil {
			return nil, err, BOSCLI_EMPTY_CODE
		}
		return &putBucketAclArgs{bucketName: bucketName, acl: ruleJosn, opType: 1}, nil, BOSCLI_OK
	}

	return nil, nil, BOSCLI_PUT_ACL_CANNED_FILE_BOTH_EMPTY
}

// Executing put acl
func (b *BosApi) putBucketAclExecute(opType int, aclJosn []byte, bucketName, canned string) (error,
	BosCliErrorCode) {

	// put canned ACL
	if opType == 2 {
		if err := b.bosClient.PutBucketAclFromCanned(bucketName, canned); err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
	} else {
		// print acl from file.
		var out bytes.Buffer
		json.Indent(&out, aclJosn, "", "  ")
		out.WriteTo(os.Stdout)

		// put ACL
		if err := b.bosClient.PutBucketAclFromString(bucketName, string(aclJosn)); err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
	}
	return nil, BOSCLI_OK
}

// Get ACL
// must have bucket_name
func (b *BosApi) GetBucketAcl(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getBucketAclPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	if err := b.getBucketAclExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Get ACL preprocessing
func (b *BosApi) getBucketAclPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Get ACL execute
func (b *BosApi) getBucketAclExecute(bucketName string) error {
	// get Acl
	ret, err := b.bosClient.GetBucketAcl(bucketName)
	if err != nil {
		return err
	}

	// print ACL
	aclJosn, err := json.Marshal(ret)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, aclJosn, "", "  ")
	fmt.Println(out.String())
	return nil
}

type putLifecycleArgs struct {
	bucketName string
	lifecycle  []byte
}

// Put life cycle
// must have bucket_name
func (b *BosApi) PutLifecycle(lifecycleConfigPath, bosPath string, template bool) {

	// preprocessing
	args, err, retCode := b.putLifecyclePreProcess(lifecycleConfigPath, bosPath, template)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}

	//executing
	err, retCode = b.putLifecycleExecute(args.lifecycle, args.bucketName, template)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}
}

// Put lifecycle preprocessing
func (b *BosApi) putLifecyclePreProcess(lifecycleConfigPath, bosPath string,
	template bool) (*putLifecycleArgs, error, BosCliErrorCode) {
	// show template?
	if template {
		return &putLifecycleArgs{}, nil, BOSCLI_OK
	}

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return nil, nil, retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return nil, nil, BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return nil, nil, BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	// have config path and bucket name?
	if lifecycleConfigPath == "" || bucketName == "" {
		return nil, fmt.Errorf("you must specify lifecycle_config_file and bucket_name"),
			BOSCLI_PUT_LIFECYCLE_NO_CONFIG_AND_BUCKET
	}

	// config path exist?
	if !util.DoesFileExist(lifecycleConfigPath) {
		return nil, nil, boscmd.LOCAL_FILE_NOT_EXIST
	}

	// lifecycle config is valid?
	fd, err := os.Open(lifecycleConfigPath)
	if err != nil {
		return nil, err, BOSCLI_EMPTY_CODE
	}
	defer fd.Close()

	ruleJosn, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err, BOSCLI_EMPTY_CODE
	}

	rule := &api.GetBucketLifecycleResult{}
	lifecycleRule := json.NewDecoder(bytes.NewReader(ruleJosn))
	if err := lifecycleRule.Decode(rule); err != nil {
		return nil, err, BOSCLI_EMPTY_CODE
	}

	return &putLifecycleArgs{
		bucketName: bucketName,
		lifecycle:  ruleJosn,
	}, nil, BOSCLI_OK
}

// Executing put lifecycle
func (b *BosApi) putLifecycleExecute(ruleJosn []byte, bucketName string, template bool) (error,
	BosCliErrorCode) {

	if template {
		// print template
		ruleTemplate := api.GetBucketLifecycleResult{
			Rule: []api.LifecycleRuleType{
				api.LifecycleRuleType{
					Id:       "sample-id",
					Status:   "enabled",
					Resource: []string{"${bucket_name}/${prefix}/*"},
					Condition: api.LifecycleConditionType{
						Time: api.LifecycleConditionTimeType{
							DateGreaterThan: "$(lastModified)+P30D",
						},
					},
					Action: api.LifecycleActionType{
						Name:         "Transition",
						StorageClass: "STANDARD_IA",
					},
				},
			},
		}
		ruleJosn, err := json.Marshal(ruleTemplate)
		if err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
		var out bytes.Buffer
		json.Indent(&out, ruleJosn, "", "  ")
		out.WriteTo(os.Stdout)
	} else {
		// print lifecycle.
		var out bytes.Buffer
		json.Indent(&out, ruleJosn, "", "  ")
		out.WriteTo(os.Stdout)

		// put lifecycle
		if err := b.bosClient.PutBucketLifecycleFromString(bucketName,
			string(ruleJosn)); err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
	}
	return nil, BOSCLI_OK
}

// Get life cycle
// must have bucket_name
func (b *BosApi) GetLifecycle(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getLifecyclePreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// get lifecycle
	if err := b.getLifecycleExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Get lifecycle reprocessing
func (b *BosApi) getLifecyclePreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Get lifecycle execute
func (b *BosApi) getLifecycleExecute(bucketName string) error {
	ret, err := b.bosClient.GetBucketLifecycle(bucketName)
	if err != nil {
		return err
	}

	// print lifecycle
	lifecycleJosn, err := json.Marshal(ret)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, lifecycleJosn, "", "  ")
	fmt.Println(out.String())

	return nil
}

// Delete life cycle
// must have bucket_name
func (b *BosApi) DeleteLifecycle(bosPath string) {
	// check bucket name
	bucketName, retCode := b.deleteLifecyclePreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// delete lifecycle
	if err := b.deleteLifecycleExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}

}

// delete lifecycle reprocessing
func (b *BosApi) deleteLifecyclePreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Delete lifecycle execute
func (b *BosApi) deleteLifecycleExecute(bucketName string) error {
	return b.bosClient.DeleteBucketLifecycle(bucketName)
}

// Put logging
// must have bucket_name
func (b *BosApi) PutLogging(targetBosPath, targetPrefix, bosPath string) {

	// check request
	bucketName, targetName, retCode := b.putLoggingPreProcess(targetBosPath, targetPrefix, bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// start put logging
	if err := b.putLoggingExecute(targetName, targetPrefix, bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Put logging preprocessing
func (b *BosApi) putLoggingPreProcess(targetBosPath, targetPrefix,
	bosPath string) (string, string, BosCliErrorCode) {

	// preprocess bos path
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	// preprocess target path
	retCode, err = checkBosPath(targetBosPath)
	if err != nil {
		return "", "", retCode
	}
	targetName, objectKey := splitBosBucketKey(targetBosPath)
	if targetName == "" {
		return "", "", BOSCLI_PUT_LOG_NO_TARGET_BUCKET
	} else if objectKey != "" {
		return "", "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, targetName, BOSCLI_OK
}

// Executing put logging
func (b *BosApi) putLoggingExecute(targetName, targetPrefix, bucketName string) error {
	// print logging infomation
	loggingString := "{\n" +
		"  \"status\": \"enabled\"" + "\n" +
		"  \"targetPrefix\": \"" + targetPrefix + "\"\n" +
		"  \"targetBucket\": \"" + targetName + "\"\n" +
		"}"
	fmt.Println(loggingString)

	// put logging configuration to bos
	args := &api.PutBucketLoggingArgs{
		TargetBucket: targetName,
		TargetPrefix: targetPrefix,
	}
	return b.bosClient.PutBucketLoggingFromStruct(bucketName, args)
}

// Get logging
// must have bucket_name
func (b *BosApi) GetLogging(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getLoggingPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// get logging
	if err := b.getLoggingExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Get logging reprocessing
func (b *BosApi) getLoggingPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Get logging Execute
func (b *BosApi) getLoggingExecute(bucketName string) error {
	ret, err := b.bosClient.GetBucketLogging(bucketName)
	if err != nil {
		return err
	}

	// print logging information
	loggingJosn, err := json.Marshal(ret)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, loggingJosn, "", "  ")
	fmt.Println(out.String())

	return nil
}

// Delete logging
// must have bucket_name
func (b *BosApi) DeleteLogging(bosPath string) {
	// check bucket name
	bucketName, retCode := b.deleteLoggingPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// delete logging
	if err := b.deleteLoggingExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Delete logging reprocessing
func (b *BosApi) deleteLoggingPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Delete logging execute
func (b *BosApi) deleteLoggingExecute(bucketName string) error {
	return b.bosClient.DeleteBucketLogging(bucketName)
}

// Put storage class
// must have bucket-name and storage-class
func (b *BosApi) PutBucketStorageClass(bosPath, storageClass string) {
	// check request
	bucketName, retCode := b.putBucketStorageClassPreProcess(bosPath, storageClass)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// put storage class
	if err := b.putBucketStorageClassExecute(bucketName, storageClass); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// put storage class preprocessing
func (b *BosApi) putBucketStorageClassPreProcess(bosPath, storageClass string) (string,
	BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	// check storage class is valid
	if storageClass == "" {
		return "", BOSCLI_STORAGE_CLASS_IS_EMPTY
	}
	if _, retCode := getStorageClassFromStr(storageClass); retCode != BOSCLI_OK {
		return "", retCode
	}
	return bucketName, BOSCLI_OK
}

// Put storage class execute
func (b *BosApi) putBucketStorageClassExecute(bucketName, storageClass string) error {
	return b.bosClient.PutBucketStorageclass(bucketName, storageClass)
}

// Get storage class
// must have bucket-name and storage-class
func (b *BosApi) GetBucketStorageClass(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getBucketStorageClassPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// get storage class
	err := b.getBucketStorageClassExecute(bucketName)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
}

func (b *BosApi) GetObjectMeta(bucketName string, objectName string) (*api.GetObjectMetaResult, error) {
	return b.bosClient.GetObjectMeta(bucketName, objectName)
}

// get storage class preprocessing
func (b *BosApi) getBucketStorageClassPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

func (b *BosApi) getBucketStorageClassExecute(bucketName string) error {
	// get storage class
	ret, err := b.bosClient.GetBucketStorageclass(bucketName)
	if err != nil {
		return err
	}

	// print logging information
	fmt.Printf("{\n    \"storageClass\": \"%s\"\n}\n", ret)

	return nil
}
