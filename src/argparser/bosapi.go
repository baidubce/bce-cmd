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

package argparser

import (
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"log"
)

import (
	"bcecmd/boscmd/boscli"
)

var (
	bosapiClient *boscli.BosApi
)

// get a new bosapi client
func initBosapiClient() {
	if bosapiClient != nil {
		return
	}
	bosapiClient = boscli.NewBosApi()
}

type BosApiArgs struct {
	srcBosPath    string
	dstBosPath    string
	srcBosKeyPath string
	srcPath       string
	storageClass  string
	template      bool
	canned        string
}

// Put ACL
func (b *BosApiArgs) putBucketAcl(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.PutBucketAcl(b.srcPath, b.srcBosPath, b.canned)
	return nil
}

// Get ACL
func (b *BosApiArgs) getBucketAcl(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.GetBucketAcl(b.srcBosPath)
	return nil
}

// Put lifecycle
func (b *BosApiArgs) putLifecycle(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.PutLifecycle(b.srcPath, b.srcBosPath, b.template)
	return nil
}

// Get lifecycle
func (b *BosApiArgs) getLifecycle(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.GetLifecycle(b.srcBosPath)
	return nil
}

// Delete lifecycle
func (b *BosApiArgs) deleteLifecycle(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.DeleteLifecycle(b.srcBosPath)
	return nil
}

// Put logging
func (b *BosApiArgs) putLogging(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.PutLogging(b.srcBosPath, b.srcBosKeyPath, b.dstBosPath)
	return nil
}

// Get logging
func (b *BosApiArgs) getLogging(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.GetLogging(b.srcBosPath)
	return nil
}

// Delete logging
func (b *BosApiArgs) deleteLogging(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.DeleteLogging(b.srcBosPath)
	return nil
}

// Put storage class
func (b *BosApiArgs) putBucketStorageClass(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.PutBucketStorageClass(b.srcBosPath, b.storageClass)
	return nil
}

// Get storage class
func (b *BosApiArgs) getBucketStorageClass(context *kingpin.ParseContext) error {
	initBosapiClient()
	bosapiClient.GetBucketStorageClass(b.srcBosPath)
	return nil
}

func (b *BosApiArgs) getObjectMeta(context *kingpin.ParseContext) error {
	initBosapiClient()
	result, err := bosapiClient.GetObjectMeta(b.srcBosPath, b.srcBosKeyPath)
	if err != nil {
		return err
	}
	tempByte, _ := json.Marshal(*result)
	log.Println(string(tempByte))
	return nil
}

// build parser for put acl
func buildPutBucketAclParser(putBucketAclCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	putBucketAclCmd.Action(bosApiArgsValue.putBucketAcl)
	putBucketAclCmd.Flag(
		"acl-config-file",
		"path to acl file in json format.").
		StringVar(&bosApiArgsValue.srcPath)
	putBucketAclCmd.Flag(
		"bucket-name",
		"bucket you want to put acl for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
	putBucketAclCmd.Flag(
		"canned",
		"set the canned acl of the given bucket, it can be: 'private', 'public-read' or "+
			"'public-read-write'").
		StringVar(&bosApiArgsValue.canned)
}

// build parser for get acl
func buildGetBucketAclParser(getBucketAclCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	getBucketAclCmd.Action(bosApiArgsValue.getBucketAcl)
	getBucketAclCmd.Flag(
		"bucket-name",
		"bucket you want to get acl for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

// build parser for put lifecycle
func buildPutLifecycleParser(putLifecycleCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	putLifecycleCmd.Action(bosApiArgsValue.putLifecycle)
	putLifecycleCmd.Flag(
		"lifecycle-config-file",
		"path to lifecycle file in json format, use --template to get an template of the file.").
		StringVar(&bosApiArgsValue.srcPath)
	putLifecycleCmd.Flag(
		"bucket-name",
		"bucket you want to put lifecycle for.").
		StringVar(&bosApiArgsValue.srcBosPath)
	putLifecycleCmd.Flag(
		"template",
		"generates a lifeycle template for the given bucket.").
		BoolVar(&bosApiArgsValue.template)
}

// build parser for get life cycle
func buildGetLifecycleParser(getLifecycleCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	getLifecycleCmd.Action(bosApiArgsValue.getLifecycle)
	getLifecycleCmd.Flag(
		"bucket-name",
		"bucket you want to get lifecycle config for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

// build parser for delete life cycle
func buildDelLifecycleParser(delLifecycleCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	delLifecycleCmd.Action(bosApiArgsValue.deleteLifecycle)
	delLifecycleCmd.Flag(
		"bucket-name",
		"bucket you want to delete lifecycle config.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

// build parser for put logging parser
func putLoggingParser(putLoggingCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	putLoggingCmd.Action(bosApiArgsValue.putLogging)
	putLoggingCmd.Flag(
		"target-bucket",
		"which bucket will the log put to.").
		StringVar(&bosApiArgsValue.srcBosPath)
	putLoggingCmd.Flag(
		"target-prefix",
		"which prefix will the log put to.").
		StringVar(&bosApiArgsValue.srcBosKeyPath)
	putLoggingCmd.Flag(
		"bucket-name",
		"bucket you want to put logging for.").
		Required().
		StringVar(&bosApiArgsValue.dstBosPath)
}

// build parser for get logging
func getLoggingParser(getLoggingCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	getLoggingCmd.Action(bosApiArgsValue.getLogging)
	getLoggingCmd.Flag(
		"bucket-name",
		"bucket you want to list logging for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

// build parser for delete logging
func delLoggingParser(delLoggingCmd *kingpin.CmdClause, bosApiArgsValue *BosApiArgs) {
	delLoggingCmd.Action(bosApiArgsValue.deleteLogging)
	delLoggingCmd.Flag(
		"bucket-name",
		"bucket you want to delete logging for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

// build parser for put storage class
func putBucketStorageClassParser(putBucketStorageClassCmd *kingpin.CmdClause,
	bosApiArgsValue *BosApiArgs) {
	putBucketStorageClassCmd.Action(bosApiArgsValue.putBucketStorageClass)
	putBucketStorageClassCmd.Flag(
		"bucket-name",
		"bucket you want to put bucket storege for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
	putBucketStorageClassCmd.Flag(
		"storage-class",
		"bucket storage class, should be STANDARD or STANDARD_IA or COLD.").
		Required().
		StringVar(&bosApiArgsValue.storageClass)
}

// build parser for storage class
func getBucketStorageClassParser(getBucketStorageClassCmd *kingpin.CmdClause,
	bosApiArgsValue *BosApiArgs) {
	getBucketStorageClassCmd.Action(bosApiArgsValue.getBucketStorageClass)
	getBucketStorageClassCmd.Flag(
		"bucket-name",
		"bucket you want to put bucket storege for.").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
}

func getObjectMetaParser(getObjectMetaCmd *kingpin.CmdClause,
	bosApiArgsValue *BosApiArgs) {
	getObjectMetaCmd.Action(bosApiArgsValue.getObjectMeta)
	getObjectMetaCmd.Flag(
		"bucket-name",
		"bucket name you want to get object from").
		Required().
		StringVar(&bosApiArgsValue.srcBosPath)
	getObjectMetaCmd.Flag(
		"object-name",
		"object name you want to get").Required().StringVar(&bosApiArgsValue.srcBosKeyPath)
}

// BOS API argparser builder
func BuildBosApi(bosApi *kingpin.CmdClause) {
	bosApiArgsValue := &BosApiArgs{}

	putBucketAclCmd := bosApi.Command("put-bucket-acl", "put bucket ACL.")
	buildPutBucketAclParser(putBucketAclCmd, bosApiArgsValue)

	getBucketAclCmd := bosApi.Command("get-bucket-acl", "get bucket ACL.")
	buildGetBucketAclParser(getBucketAclCmd, bosApiArgsValue)

	putLifecycleCmd := bosApi.Command("put-lifecycle", "put lifecycle.")
	buildPutLifecycleParser(putLifecycleCmd, bosApiArgsValue)

	getLifecycleCmd := bosApi.Command("get-lifecycle", "get lifecycle.")
	buildGetLifecycleParser(getLifecycleCmd, bosApiArgsValue)

	delLifecycleCmd := bosApi.Command("delete-lifecycle", "delete lifecycle.")
	buildDelLifecycleParser(delLifecycleCmd, bosApiArgsValue)

	putLoggingCmd := bosApi.Command("put-logging", "put logging.")
	putLoggingParser(putLoggingCmd, bosApiArgsValue)

	getLoggingCmd := bosApi.Command("get-logging", "get logging.")
	getLoggingParser(getLoggingCmd, bosApiArgsValue)

	delLoggingCmd := bosApi.Command("delete-logging", "delete logging.")
	delLoggingParser(delLoggingCmd, bosApiArgsValue)

	putBucketStorageClassCmd := bosApi.Command("put-bucket-storage-class",
		"storage class configuration, should be STANDARD or STANDARD_IA or COLD.")
	putBucketStorageClassParser(putBucketStorageClassCmd, bosApiArgsValue)

	getBucketStorageClassCmd := bosApi.Command("get-bucket-storage-class",
		"get bucket storage class.")
	getBucketStorageClassParser(getBucketStorageClassCmd, bosApiArgsValue)

	getObjectMetaCmd := bosApi.Command("get-object-meta", "get object meta")
	getObjectMetaParser(getObjectMetaCmd, bosApiArgsValue)

}
