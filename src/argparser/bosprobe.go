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
	"bcecmd/boscmd/bosprobe"
)

// bosProbeArgs is used to store the arguments of bosprobe
type bosProbeArgs struct {
	ak         string
	sk         string
	url        string
	bucketName string
	objectName string
	localPath  string
	endpoint   string
}

// wrapper function of bosprobe upload check
func (b *bosProbeArgs) probeUploadCheck(context *kingpin.ParseContext) error {
	req := &bosprobe.UploadCheckRes{
		Ak:         b.ak,
		Sk:         b.sk,
		BucketName: b.bucketName,
		ObjectName: b.objectName,
		LocalPath:  b.localPath,
		Endpoint:   b.endpoint,
	}
	c := &bosprobe.UploadCheck{}
	bosprobe.Handler(c, req)
	return nil
}

// wrapper function of bosprobe download check
func (b *bosProbeArgs) probeDownloadCheck(context *kingpin.ParseContext) error {
	req := &bosprobe.DownloadCheckRes{
		Ak:         b.ak,
		Sk:         b.sk,
		FileUrl:    b.url,
		BucketName: b.bucketName,
		ObjectName: b.objectName,
		LocalPath:  b.localPath,
		Endpoint:   b.endpoint,
	}
	c := &bosprobe.DownloadCheck{}
	bosprobe.Handler(c, req)
	return nil
}

// build parser for bosprobe
func BuildBosProbeParser(bosProbeCmd *kingpin.CmdClause) {
	probeArgsVal := &bosProbeArgs{}

	probeUploadCmd := bosProbeCmd.Command("upload", "上传检测").Action(
		probeArgsVal.probeUploadCheck)
	probeUploadCmd.Flag(
		"ak",
		"您的AK（非必需，如果测试的Bucket为公共读写，则不需要配置）.").
		Short('a').StringVar(&probeArgsVal.ak)

	probeUploadCmd.Flag(
		"sk",
		"您的SK（非必需，如果测试的Bucket为公共读写，则不需要配置）.").
		Short('s').StringVar(&probeArgsVal.sk)

	probeUploadCmd.Flag(
		"from",
		"指定你要上传的文件（非必需，如果您不指定，我们会随机生成大小为100MB的文件）").
		Short('f').StringVar(&probeArgsVal.localPath)

	probeUploadCmd.Flag(
		"bucket",
		"bucket 名称").
		Short('b').Required().StringVar(&probeArgsVal.bucketName)

	probeUploadCmd.Flag(
		"object",
		"文件在bucket中保存的名称（非必需，如果不指定，我们将使用本地文件的名称）").
		Short('o').
		StringVar(&probeArgsVal.objectName)

	probeUploadCmd.Flag(
		"endpoint",
		"你可以指定endpoint （非必需，如果不指定，我们将根据bucket名字自动指定endpoint）").
		Short('e').StringVar(&probeArgsVal.endpoint)

	probeDownloadCmd := bosProbeCmd.Command(
		"download",
		"下载检测: 你可以通过-f 指定源文件地址，也可通过-b和-o组合指定源文件地址，"+
			"但是‘-f’和‘ -b -o’ 不能同时使用。").
		Action(probeArgsVal.probeDownloadCheck)
	probeDownloadCmd.Flag(
		"ak",
		"您的AK（非必需，如果测试的Bucket为公共读写，则不需要配置）.").
		Short('a').StringVar(&probeArgsVal.ak)

	probeDownloadCmd.Flag(
		"sk",
		"您的SK（非必需，如果测试的Bucket为公共读写，则不需要配置）.").
		Short('s').StringVar(&probeArgsVal.sk)

	probeDownloadCmd.Flag(
		"from",
		"指定你要下载文件的URL（非必需，您可以指定bucket name 和 object name）").
		Short('f').StringVar(&probeArgsVal.url)

	probeDownloadCmd.Flag(
		"bucket",
		"bucket 名称").
		Short('b').StringVar(&probeArgsVal.bucketName)

	probeDownloadCmd.Flag(
		"object",
		"文件在bucket中保存的名称（非必需， 如果不指定，我们将使用本地文件的名称）.").
		Short('o').StringVar(&probeArgsVal.objectName)

	probeDownloadCmd.Flag(
		"to",
		"你可以指定下载文件保存的名称（非必需，如果不指定，我们将自动指定object名称）.").
		Short('t').StringVar(&probeArgsVal.localPath)

	probeDownloadCmd.Flag(
		"endpoint",
		"你可以指定endpoint（非必需，如果不指定，我们将根据bucket名字自动指定endpoint）.\n"+
			"\n"+
			"例:\n"+
			"1. 不带endpoint:\n"+
			"bcecmd bosprobe download –a 123 –s 456 –b bucket-bj –o file1.txt –t ./tmp/example.txt\n"+
			"2. 指定endpoint为bj.bcebos.com:\n"+
			"bcecmd bosprobe download –a 123 –s 456 –b bucket-bj –o file1.txt –t ./tmp/example.txt "+
			"-e bj.bcebos.com\n"+
			"\n"+
			"说明：\n"+
			"endpoint 与你使用的bucket所在区域有关，区域与endpoint的对应关系为：\n"+
			"|------|---------------|\n"+
			"| 北京 | bj.bcebos.com  |\n"+
			"| 苏州 | su.bcebos.com  |\n"+
			"| 广州 | gz.bcebos.com  |\n"+
			"|香港2 |hk-2.bcebos.com|\n"+
			"|------|---------------|\n"+
			"\n"+
			"注意: \n"+
			"endpoint不能带http或者https\n").
		Short('e').StringVar(&probeArgsVal.endpoint)
}
