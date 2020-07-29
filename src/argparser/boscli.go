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
	"github.com/alecthomas/kingpin"
)

import (
	"bcecmd/boscmd/boscli"
)

const (
	EXPIRES_VAL_FOR_NOT_SET = -123213348
)

var (
	boscliClient *boscli.BosCli
)

func initBoscliClient() {
	if boscliClient != nil {
		return
	}
	boscliClient = boscli.NewBosCli()
}

type BosArgs struct {
	bosPath       string
	srcPath       string
	dstPath       string
	storageClass  string
	syncType      string
	region        string
	downLoadTmp   string
	exclude       []string
	include       []string
	excludeTime   []string
	includeTime   []string
	excludeDelete []string
	expires       int
	concurrency   int
	all           bool
	recursive     bool
	summerize     bool
	restart       bool
	force         bool
	yes           bool
	dryrun        bool
	del           bool
	quiet         bool
	disableBar    bool
}

// Gen signed url
func (b *BosArgs) genSignedUrl(context *kingpin.ParseContext) error {
	initBoscliClient()
	// when expires is 0, it maybe be set by user or the default value of int
	haveSet := false
	if b.expires != EXPIRES_VAL_FOR_NOT_SET {
		haveSet = true
	}
	boscliClient.GenSignedUrl(b.bosPath, b.expires, haveSet)
	return nil
}

// list buckets or objects
func (b *BosArgs) bosList(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.List(b.bosPath, b.all, b.recursive, b.summerize)
	return nil
}

// make bucket
func (b *BosArgs) makeBucket(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.MakeBucket(b.bosPath, b.region, b.quiet)
	return nil
}

// remove bucket
func (b *BosArgs) rmoveBucket(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.RemoveBucket(b.bosPath, b.force, b.yes, b.quiet)
	return nil
}

// remove objects
func (b *BosArgs) rmoveObject(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.RemoveObject(b.bosPath, b.yes, b.recursive, b.quiet)
	return nil
}

// upload, download or copy objects
func (b *BosArgs) bosCopy(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.Copy(b.srcPath, b.dstPath, b.storageClass, b.downLoadTmp, b.recursive, b.restart, b.quiet,
		b.yes, b.disableBar)
	return nil
}

// sync
func (b *BosArgs) bosSync(context *kingpin.ParseContext) error {
	initBoscliClient()
	boscliClient.Sync(b.srcPath, b.dstPath, b.storageClass, b.downLoadTmp, b.syncType, b.exclude, b.include,
		b.excludeTime, b.includeTime, b.excludeDelete, b.concurrency, b.del, b.dryrun, b.yes, b.quiet, true,
		b.restart)
	return nil
}

// build parser for generate signed url
func buildGenParser(genCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {
	bosArgsValue.expires = EXPIRES_VAL_FOR_NOT_SET
	genCmd.Action(bosArgsValue.genSignedUrl)
	genCmd.Arg(
		"BOS_PATH",
		"the BOS path will be used by signed url.").
		Required().StringVar(&bosArgsValue.bosPath)
	genCmd.Flag(
		"expires",
		"you can specify the expiration time for the signed url, the expiration"+
			"time must be equal or greater than -1, in which -1 means never expires.").
		Short('e').IntVar(&bosArgsValue.expires)
}

// build parser for list buckets and list objects
func buildLsParser(lsCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {

	lsCmd.Action(bosArgsValue.bosList)
	lsCmd.Arg(
		"BOS_PATH",
		"BOS path start with \"bos:/\". only 1000 objects would be listed "+
			"if there are more objects in a bucket. ").
		Default("bos:/").StringVar(&bosArgsValue.bosPath)

	lsCmd.Flag(
		"all",
		"list all objects and subdirs.").
		Short('a').BoolVar(&bosArgsValue.all)

	lsCmd.Flag(
		"recursive",
		"list objects under subdirs").
		Short('r').BoolVar(&bosArgsValue.recursive)

	lsCmd.Flag(
		"summerize",
		"show summerization").
		Short('s').BoolVar(&bosArgsValue.summerize)
}

// build parser for make bucket
func buildMbParser(mbCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {
	mbCmd.Action(bosArgsValue.makeBucket)
	mbCmd.Arg(
		"BUCKET_NAME",
		"bucket name you want to create.").
		Required().StringVar(&bosArgsValue.bosPath)
	mbCmd.Flag(
		"region",
		"specify the region for the bucket").
		Short('r').StringVar(&bosArgsValue.region)
	mbCmd.Flag(
		"quiet",
		"do not display the operations performed from the specified command").
		BoolVar(&bosArgsValue.quiet)
}

// build parser for remove bucket
func buildRbParser(rbCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {
	rbCmd.Action(bosArgsValue.rmoveBucket)
	rbCmd.Arg(
		"BUCKET_NAME",
		"bucket name you want to delete.").
		Required().StringVar(&bosArgsValue.bosPath)
	rbCmd.Flag(
		"force",
		"delete bucket and ALL objects in it.").
		Short('f').BoolVar(&bosArgsValue.force)
	rbCmd.Flag(
		"yes",
		"delete bucket without any prompt").
		Short('y').BoolVar(&bosArgsValue.yes)
	rbCmd.Flag(
		"quiet",
		"do not display the operations performed from the specified command").
		BoolVar(&bosArgsValue.quiet)
}

// build parser for remove object
func buildRmParser(rmCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {
	rmCmd.Action(bosArgsValue.rmoveObject)
	rmCmd.Arg(
		"BOS_PATH",
		"BOS path start with \"bos:/\"").
		Required().StringVar(&bosArgsValue.bosPath)
	rmCmd.Flag(
		"recursive",
		"delete objects under subdirs").
		Short('r').BoolVar(&bosArgsValue.recursive)
	rmCmd.Flag(
		"yes",
		"delete objects without any prompt").
		Short('y').BoolVar(&bosArgsValue.yes)
	rmCmd.Flag(
		"quiet",
		"do not display the operations performed from the specified command").
		BoolVar(&bosArgsValue.quiet)
}

// build parser for copy
func buildCopyParser(cpCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {

	cpCmd.Action(bosArgsValue.bosCopy)
	cpCmd.Arg(
		"SRC",
		"source path, could be either local or BOS path. When source path, "+
			"could be either local or BOS path.").
		Required().StringVar(&bosArgsValue.srcPath)

	cpCmd.Arg(
		"DST",
		"destination path, could be either local or BOS path. ").
		Required().StringVar(&bosArgsValue.dstPath)

	cpCmd.Flag(
		"recursive",
		"list objects under subdirs").
		Short('r').BoolVar(&bosArgsValue.recursive)

	cpCmd.Flag(
		"restart",
		"restart upload object.").
		BoolVar(&bosArgsValue.restart)

	cpCmd.Flag(
		"storage-class",
		"storage class configuration, should be STANDARD or STANDARD_IA or COLD").
		StringVar(&bosArgsValue.storageClass)

	cpCmd.Flag(
		"download-tmp-path",
		"the path of temporary folder that stores temporary files for breakpoint downloading").
		StringVar(&bosArgsValue.downLoadTmp)

	cpCmd.Flag(
		"quiet",
		"do not display the operations performed from the specified command").
		BoolVar(&bosArgsValue.quiet)

	cpCmd.Flag(
		"yes",
		"without any prompt").
		Short('y').BoolVar(&bosArgsValue.yes)

	cpCmd.Flag(
		"disable-bar",
		"not display progress bar").
		BoolVar(&bosArgsValue.disableBar)
}

// build parser for sync
func buildSyncParser(syncCmd *kingpin.CmdClause, bosArgsValue *BosArgs) {

	syncCmd.Action(bosArgsValue.bosSync)
	syncCmd.Arg(
		"SRC",
		"source path, should be BOS path or local path.").
		Required().StringVar(&bosArgsValue.srcPath)

	syncCmd.Arg(
		"DST",
		"destination path, should be BOS path or local path.").
		Required().StringVar(&bosArgsValue.dstPath)

	syncCmd.Flag(
		"exclude",
		"multiple patterns to filter file when sync; the value should be quoted if it contains "+
			"wildcard(*). e.g: \n"+
			"--exclude './.svn/*'; \n"+
			"--exclude ./path/to/file; \n"+
			"--exclude '*/file'; \n"+
			"--exclude '*.jpg';\n "+
			"--exclude 'bos:/bucket/path/*'\n"+
			"*NOTE:* In order to exclude an entire folder, the pattern must end with wildcard! "+
			"Such as --exclude 'dir/*'.").
		StringsVar(&bosArgsValue.exclude)

	syncCmd.Flag(
		"include",
		"multiple patterns to specify the files that needed to synchronized; the value should be "+
			"quoted if it contains. e.g:\n "+
			"--include './.svn/*';\n"+
			"--include ./path/to/file;\n"+
			"--include '*/file';\n"+
			"--include '*.jpg';\n"+
			"--include 'bos:/bucket/path/*'\n").
		StringsVar(&bosArgsValue.include)

	syncCmd.Flag(
		"delete",
		"delete objects of destination which do not exist in the source").
		BoolVar(&bosArgsValue.del)

	syncCmd.Flag(
		"exclude-delete",
		"multiple patterns to filter file when delete objects of destination which do not exist "+
			"in the source; the value should be quoted if it contains "+
			"wildcard(*). e.g: \n"+
			"--exclude-delete './.svn/*'; \n"+
			"--exclude-delete ./path/to/file; \n"+
			"--exclude-delete '*/file'; \n"+
			"--exclude-delete '*.jpg';\n "+
			"--exclude-delete 'bos:/bucket/path/*'\n"+
			"*NOTE:* In order to exclude an entire folder, the pattern must end with wildcard! "+
			"Such as --exclude 'dir/*'.").
		StringsVar(&bosArgsValue.excludeDelete)

	syncCmd.Flag(
		"dryrun",
		"list what will be synced, and what will be deleted(if --delete is specified)").
		BoolVar(&bosArgsValue.dryrun)

	syncCmd.Flag(
		"yes",
		"continue doing things with positive confirmations").
		BoolVar(&bosArgsValue.yes)

	syncCmd.Flag(
		"quiet",
		"do not display the operations performed from the specified command").
		BoolVar(&bosArgsValue.quiet)

	syncCmd.Flag(
		"storage-class",
		"storage class configuration, should be STANDARD or STANDARD_IA or COLD").
		StringVar(&bosArgsValue.storageClass)

	syncCmd.Flag(
		"sync-type",
		"sync-type should be 'time-size' or 'time-size-crc32' or 'only-crc32', the default is "+
			" 'time-size':\n"+
			"  time-size: bcecmd will sync a file when the modification time of the destination "+
			"    is early than source or the modification time is the same and the size is "+
			"differernt; \n"+
			"  time-size-crc32: bcecmd will compare then modification time and size of "+
			"    the same name files as the time-size mode, but if the result of time-size mode is"+
			"    need sync this file,  bcecmd will compare the crc32 of this file;\n"+
			"	only-crc32: bcecmd only compare the crc32 of the same name files.\n"+
			"*NOTE:* if bcecmd can't get the crc32 of objects from BOS, bcecmd will away sync "+
			"these objects (BOS don't store the crc32 of old objects which are uploaded to BOS "+
			"before 2018-01),  ").
		StringVar(&bosArgsValue.syncType)

	syncCmd.Flag(
		"download-tmp-path",
		"the path of temporary folder that stores temporary files for breakpoint downloading").
		StringVar(&bosArgsValue.downLoadTmp)

	syncCmd.Flag(
		"concurrency",
		"max concurrency for sync, default value is multi upload number").
		IntVar(&bosArgsValue.concurrency)

	syncCmd.Flag(
		"restart",
		"don't transfer from breakpoint.").
		BoolVar(&bosArgsValue.restart)
}

func BuildBosParser(bos *kingpin.CmdClause) {
	bosArgsValue := &BosArgs{}

	genCmd := bos.Command("gen_signed_url", "generate signed url with given BOS path.")
	buildGenParser(genCmd, bosArgsValue)

	lsCmd := bos.Command("ls", "list buckets or objects.").Alias("list")
	buildLsParser(lsCmd, bosArgsValue)

	cpCmd := bos.Command("cp", "copy objects among local and BOS.").Alias("copy")
	buildCopyParser(cpCmd, bosArgsValue)

	mbCmd := bos.Command("mb", "make bucket.").Alias("make-bucket")
	buildMbParser(mbCmd, bosArgsValue)

	rbCmd := bos.Command("rb", "remove bucket.").Alias("remove-bucket")
	buildRbParser(rbCmd, bosArgsValue)

	rmCmd := bos.Command("rm", "remove objects.").Alias("remove-object")
	buildRmParser(rmCmd, bosArgsValue)

	syncCmd := bos.Command("sync", "synchronize objects between local and BOS or between BOS and "+
		"BOS.")
	buildSyncParser(syncCmd, bosArgsValue)
}
