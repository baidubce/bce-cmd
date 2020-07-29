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

// This file provide functions to init and midify bos client.

package boscli

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

import (
	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"github.com/baidubce/bce-sdk-go/util/log"
)

type retryFuncType func(*bos.Client, boscliReq, interface{}) error

type boscliReq interface {
	getBucketName() string
}

// retrHandler: wrap request and response
func retryHandler(bosClient *bos.Client, call retryFuncType, req boscliReq, resp interface{},
) error {
	var (
		err error
	)

	log.Infof("retry process handler")
	// try get endpoint from cache and execute function 'call'
	bucket := req.getBucketName()
	if ok := modifiyBosClientEndpointByBucketNameCache(bosClient, bucket); ok {
		err = call(bosClient, req, resp)
		if err == nil || !shouldRetry(err) {
			return err
		}
	}
	// retry execute function 'call', but get endpiont from BOS
	log.Debugf("First process failed or can't get endpoint of bucket from cache, error: %s", err)
	epErr := modifiyBosClientEndpointByBucketNameBos(bosClient, bucket)
	if epErr == nil {
		return call(bosClient, req, resp)
	}
	if err != nil {
		return err
	}
	return epErr
}

type bosClientWrapper struct {
	bosClient *bos.Client
}

// Wrapper head bucket
func (b *bosClientWrapper) HeadBucket(bucket string) error {
	return b.bosClient.HeadBucket(bucket)
}

// Wrapper ListBuckets - list all buckets
func (b *bosClientWrapper) ListBuckets() (*api.ListBucketsResult, error) {
	return b.bosClient.ListBuckets()
}

// Wrapper PutBucket - create a new bucket
func (b *bosClientWrapper) PutBucket(bucket string) (string, error) {
	return b.bosClient.PutBucket(bucket)
}

// Wrapper GetBucketLocation - get the location fo the given bucket
func (b *bosClientWrapper) GetBucketLocation(bucket string) (string, error) {
	return b.bosClient.GetBucketLocation(bucket)
}

type listObjectsReq struct {
	bucket string
	args   *api.ListObjectsArgs
}

func (l *listObjectsReq) getBucketName() string {
	return l.bucket
}

type listObjectsRsp struct {
	ret *api.ListObjectsResult
}

// Wrapper ListObjects - list all objects of the given bucket
func (b *bosClientWrapper) ListObjects(bucket string, args *api.ListObjectsArgs) (
	*api.ListObjectsResult, error) {

	req := &listObjectsReq{bucket: bucket, args: args}
	rsp := &listObjectsRsp{}

	listObjectFunc := func(bosClient *bos.Client, req boscliReq, rsp interface{}) error {
		listReq, ok := req.(*listObjectsReq)
		if !ok {
			return fmt.Errorf("Error ListObjects request type!")
		}
		listRsp, ok := rsp.(*listObjectsRsp)
		if !ok {
			return fmt.Errorf("Error ListObjects response type!")
		}
		ret, err := bosClient.ListObjects(listReq.bucket, listReq.args)
		if err == nil {
			listRsp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, listObjectFunc, req, rsp)
	if err == nil {
		return rsp.ret, nil
	}
	return nil, err
}

type olnyOneStringReq struct {
	bucket string
}

func (o *olnyOneStringReq) getBucketName() string {
	return o.bucket
}

type olnyOneStringResp struct {
	ret string
}

// Wrapper DeleteBucket - delete a empty bucket
func (b *bosClientWrapper) DeleteBucket(bucket string) error {
	req := &olnyOneStringReq{bucket: bucket}

	deleteBucketfunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		delReq, ok := req.(*olnyOneStringReq)
		if !ok {
			return fmt.Errorf("Error DeleteBucket request type!")
		}
		return bosClient.DeleteBucket(delReq.bucket)
	}

	return retryHandler(b.bosClient, deleteBucketfunc, req, nil)
}

// Wrapper BasicGeneratePresignedUrl  generate an authorization url with expire time
func (b *bosClientWrapper) BasicGeneratePresignedUrl(bucket string, object string,
	expireInSeconds int) string {

	modifiyBosClientEndpointByBucketName(b.bosClient, bucket)
	return b.bosClient.BasicGeneratePresignedUrl(bucket, object, expireInSeconds)
}

type delMultiObjectKeyListReq struct {
	keyList []string
	bucket  string
}

func (d *delMultiObjectKeyListReq) getBucketName() string {
	return d.bucket
}

type delMultiObjectKeyListResp struct {
	ret *api.DeleteMultipleObjectsResult
}

// Wrapper DeleteMultipleObjectsFromKeyList - delete a list of objects with given key string array
func (b *bosClientWrapper) DeleteMultipleObjectsFromKeyList(bucket string,
	keyList []string) (*api.DeleteMultipleObjectsResult, error) {

	req := &delMultiObjectKeyListReq{keyList: keyList, bucket: bucket}
	resp := &delMultiObjectKeyListResp{}

	delFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		delReq, ok := req.(*delMultiObjectKeyListReq)
		if !ok {
			return fmt.Errorf("Error DeleteMultipleObjectsFromKeyList request type!")
		}
		delResp, ok := resp.(*delMultiObjectKeyListResp)
		if !ok {
			return fmt.Errorf("Error DeleteMultipleObjectsFromKeyList response type!")
		}
		ret, err := bosClient.DeleteMultipleObjectsFromKeyList(delReq.bucket, delReq.keyList)
		if err == nil || err == io.EOF {
			delResp.ret = ret
			return nil
		}
		return err
	}

	err := retryHandler(b.bosClient, delFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type delSigleObjectReq struct {
	bucket string
	object string
}

func (d *delSigleObjectReq) getBucketName() string {
	return d.bucket
}

// Wrapper DeleteObject - delete the given object
func (b *bosClientWrapper) DeleteObject(bucket, object string) error {

	req := &delSigleObjectReq{bucket: bucket, object: object}

	delFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		delReq, ok := req.(*delSigleObjectReq)
		if !ok {
			return fmt.Errorf("Error DeleteObject request type!")
		}
		return bosClient.DeleteObject(delReq.bucket, delReq.object)
	}
	return retryHandler(b.bosClient, delFunc, req, nil)
}

type getObjectMetaReq struct {
	bucket string
	object string
}

func (g *getObjectMetaReq) getBucketName() string {
	return g.bucket
}

type getObjectMetaResp struct {
	ret *api.GetObjectMetaResult
}

// Wrapper GetObjectMeta
func (b *bosClientWrapper) GetObjectMeta(bucket, object string) (*api.GetObjectMetaResult, error) {
	req := &getObjectMetaReq{bucket: bucket, object: object}
	resq := &getObjectMetaResp{}

	gmFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		gmReq, ok := req.(*getObjectMetaReq)
		if !ok {
			return fmt.Errorf("Error GetObjectMeta request type!")
		}
		gmResp, ok := resp.(*getObjectMetaResp)
		if !ok {
			return fmt.Errorf("Error GetObjectMeta response type!")
		}
		ret, err := bosClient.GetObjectMeta(gmReq.bucket, gmReq.object)
		if err == nil {
			gmResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, gmFunc, req, resq)
	if err == nil {
		return resq.ret, nil
	}
	return nil, err
}

type copyObjectReq struct {
	bucket    string
	object    string
	srcBucket string
	srcObject string
	args      *api.CopyObjectArgs
}

func (c *copyObjectReq) getBucketName() string {
	return c.bucket
}

type copyObjectResp struct {
	ret *api.CopyObjectResult
}

// Wrapper Copy Object
func (b *bosClientWrapper) CopyObject(bucket, object, srcBucket, srcObject string,
	args *api.CopyObjectArgs) (*api.CopyObjectResult, error) {

	req := &copyObjectReq{
		bucket:    bucket,
		object:    object,
		srcBucket: srcBucket,
		srcObject: srcObject,
		args:      args,
	}
	resq := &copyObjectResp{}

	cbFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		cbReq, ok := req.(*copyObjectReq)
		if !ok {
			return fmt.Errorf("Error CopyObject request type!")
		}
		cbResp, ok := resp.(*copyObjectResp)
		if !ok {
			return fmt.Errorf("Error CopyObject response type!")
		}
		ret, err := bosClient.CopyObject(cbReq.bucket, cbReq.object, cbReq.srcBucket,
			cbReq.srcObject, cbReq.args)
		if err == nil {
			cbResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cbFunc, req, resq)
	if err == nil {
		return resq.ret, nil
	}
	return nil, err
}

type basicGetObjectToFileReq struct {
	bucket   string
	object   string
	filePath string
}

func (b *basicGetObjectToFileReq) getBucketName() string {
	return b.bucket
}

// Wrapper of BasiGetObjectToFile
func (b *bosClientWrapper) BasicGetObjectToFile(bucket, object, localPath string) error {
	req := &basicGetObjectToFileReq{bucket: bucket, object: object, filePath: localPath}

	bgFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		bgReq, ok := req.(*basicGetObjectToFileReq)
		if !ok {
			return fmt.Errorf("Error BasicGetObjectToFile request type!")
		}
		err := bosClient.BasicGetObjectToFile(bgReq.bucket, bgReq.object, bgReq.filePath)
		return err
	}

	return retryHandler(b.bosClient, bgFunc, req, nil)
}

type putObjectFromFileReq struct {
	bucket   string
	object   string
	fileName string
	args     *api.PutObjectArgs
}

func (p *putObjectFromFileReq) getBucketName() string {
	return p.bucket
}

type putObjectFromFileResp struct {
	ret string
}

// Wrapper of PutObjectFromFile
func (b *bosClientWrapper) PutObjectFromFile(bucket, object, fileName string,
	args *api.PutObjectArgs) (string, error) {
	req := &putObjectFromFileReq{
		bucket:   bucket,
		object:   object,
		fileName: fileName,
		args:     args,
	}
	resp := &putObjectFromFileResp{}

	poFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		poReq, ok := req.(*putObjectFromFileReq)
		if !ok {
			return fmt.Errorf("Error PutObjectFromFile request type!")
		}
		poResp, ok := resp.(*putObjectFromFileResp)
		if !ok {
			return fmt.Errorf("Error PutObjectFromFile response type!")
		}
		ret, err := bosClient.PutObjectFromFile(poReq.bucket, poReq.object, poReq.fileName,
			poReq.args)
		if err == nil {
			poResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, poFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return "", err
}

type uploadSuperFileReq struct {
	bucket       string
	object       string
	fileName     string
	storageClass string
}

func (u *uploadSuperFileReq) getBucketName() string {
	return u.bucket
}

// Wrapper of UploadSuperFile
func (b *bosClientWrapper) UploadSuperFile(bucket, object, fileName, storageClass string) error {
	req := &uploadSuperFileReq{
		bucket:       bucket,
		object:       object,
		fileName:     fileName,
		storageClass: storageClass,
	}

	usFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		usReq, ok := req.(*uploadSuperFileReq)
		if !ok {
			return fmt.Errorf("Error PutObjectFromFile request type!")
		}
		return bosClient.UploadSuperFile(usReq.bucket, usReq.object, usReq.fileName,
			usReq.storageClass)
	}

	return retryHandler(b.bosClient, usFunc, req, nil)
}

type putCannedAclReq struct {
	bucket    string
	cannedAcl string
}

func (p *putCannedAclReq) getBucketName() string {
	return p.bucket
}

// Wrapper of PutBucketAclFromCanned
func (b *bosClientWrapper) PutBucketAclFromCanned(bucket, cannedAcl string) error {
	req := &putCannedAclReq{
		bucket:    bucket,
		cannedAcl: cannedAcl,
	}

	pcFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		pcRseq, ok := req.(*putCannedAclReq)
		if !ok {
			return fmt.Errorf("Error PutBucketAclFromCanned request type!")
		}
		return bosClient.PutBucketAclFromCanned(pcRseq.bucket, pcRseq.cannedAcl)
	}

	return retryHandler(b.bosClient, pcFunc, req, nil)
}

type putAclReq struct {
	bucket string
	acl    string
}

func (p *putAclReq) getBucketName() string {
	return p.bucket
}

// Wrapper of PutBucketAcl
func (b *bosClientWrapper) PutBucketAclFromString(bucket string, acl string) error {
	req := &putAclReq{
		bucket: bucket,
		acl:    acl,
	}

	paFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		paRseq, ok := req.(*putAclReq)
		if !ok {
			return fmt.Errorf("Error PutBucketAcl request type!")
		}
		return bosClient.PutBucketAclFromString(paRseq.bucket, paRseq.acl)
	}

	return retryHandler(b.bosClient, paFunc, req, nil)
}

type getAclReq struct {
	bucket string
}

func (p *getAclReq) getBucketName() string {
	return p.bucket
}

type getAclResp struct {
	ret *api.GetBucketAclResult
}

// Wrapper of GetBucketAcl
func (b *bosClientWrapper) GetBucketAcl(bucket string) (*api.GetBucketAclResult, error) {

	req := &getAclReq{
		bucket: bucket,
	}
	resp := &getAclResp{}

	gaFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		gaReq, ok := req.(*getAclReq)
		if !ok {
			return fmt.Errorf("Error GetBucketAcl request type!")
		}
		gaResp, ok := resp.(*getAclResp)
		if !ok {
			return fmt.Errorf("Error GetBucketAcl response type!")
		}
		ret, err := bosClient.GetBucketAcl(gaReq.bucket)
		if err == nil {
			gaResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, gaFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type putLifecycleReq struct {
	bucket    string
	lifecycle string
}

func (p *putLifecycleReq) getBucketName() string {
	return p.bucket
}

// Wrapper of PutBucketLifecycleFromString
func (b *bosClientWrapper) PutBucketLifecycleFromString(bucket, lifecycle string) error {
	req := &putLifecycleReq{
		bucket:    bucket,
		lifecycle: lifecycle,
	}

	plFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		plRseq, ok := req.(*putLifecycleReq)
		if !ok {
			return fmt.Errorf("Error PutBucketLifecycle request type!")
		}
		return bosClient.PutBucketLifecycleFromString(plRseq.bucket, plRseq.lifecycle)
	}

	return retryHandler(b.bosClient, plFunc, req, nil)
}

type getLifecycleReq struct {
	bucket string
}

func (p *getLifecycleReq) getBucketName() string {
	return p.bucket
}

type getLifecycleResp struct {
	ret *api.GetBucketLifecycleResult
}

// Wrapper of GetBucketLifecycle
func (b *bosClientWrapper) GetBucketLifecycle(bucket string) (*api.GetBucketLifecycleResult,
	error) {

	req := &getLifecycleReq{
		bucket: bucket,
	}
	resp := &getLifecycleResp{}

	glFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		glReq, ok := req.(*getLifecycleReq)
		if !ok {
			return fmt.Errorf("Error GetBucketLifecycle request type!")
		}
		glResp, ok := resp.(*getLifecycleResp)
		if !ok {
			return fmt.Errorf("Error GetBucketLifecycle response type!")
		}
		ret, err := bosClient.GetBucketLifecycle(glReq.bucket)
		if err == nil {
			glResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, glFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type deleteLifecycleReq struct {
	bucket string
}

func (d *deleteLifecycleReq) getBucketName() string {
	return d.bucket
}

// Wrapper of DeleteBucketLifecycle
func (b *bosClientWrapper) DeleteBucketLifecycle(bucket string) error {

	req := &deleteLifecycleReq{
		bucket: bucket,
	}

	dlFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		dlReq, ok := req.(*deleteLifecycleReq)
		if !ok {
			return fmt.Errorf("Error DeleteBucketLifecycle request type!")
		}
		err := bosClient.DeleteBucketLifecycle(dlReq.bucket)
		return err
	}

	return retryHandler(b.bosClient, dlFunc, req, nil)
}

type putLoggingReq struct {
	bucket string
	args   *api.PutBucketLoggingArgs
}

func (p *putLoggingReq) getBucketName() string {
	return p.bucket
}

// Wrapper of PutBucketLoggingFromStruct
func (b *bosClientWrapper) PutBucketLoggingFromStruct(bucket string,
	obj *api.PutBucketLoggingArgs) error {

	req := &putLoggingReq{
		bucket: bucket,
		args:   obj,
	}

	plFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		plReq, ok := req.(*putLoggingReq)
		if !ok {
			return fmt.Errorf("Error PutBucketLoggingFromStruct request type!")
		}
		err := bosClient.PutBucketLoggingFromStruct(plReq.bucket, plReq.args)
		return err
	}

	return retryHandler(b.bosClient, plFunc, req, nil)
}

type getLoggingReq struct {
	bucket string
}

func (g *getLoggingReq) getBucketName() string {
	return g.bucket
}

type getLoggingResp struct {
	ret *api.GetBucketLoggingResult
}

// Wrapper of GetBucketLogging
func (b *bosClientWrapper) GetBucketLogging(bucket string) (*api.GetBucketLoggingResult, error) {

	req := &getLoggingReq{
		bucket: bucket,
	}
	resp := &getLoggingResp{}

	glFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		glReq, ok := req.(*getLoggingReq)
		if !ok {
			return fmt.Errorf("Error GetBucketLogging request type!")
		}
		glResp, ok := resp.(*getLoggingResp)
		if !ok {
			return fmt.Errorf("Error GetBucketLogging response type!")
		}
		ret, err := bosClient.GetBucketLogging(glReq.bucket)
		if err == nil {
			glResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, glFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type deleteLoggingReq struct {
	bucket string
}

func (d *deleteLoggingReq) getBucketName() string {
	return d.bucket
}

// Wrapper of DeleteBucketLifecycle
func (b *bosClientWrapper) DeleteBucketLogging(bucket string) error {

	req := &deleteLoggingReq{
		bucket: bucket,
	}

	dlFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		dlReq, ok := req.(*deleteLoggingReq)
		if !ok {
			return fmt.Errorf("Error DeleteBucketLogging request type!")
		}
		err := bosClient.DeleteBucketLogging(dlReq.bucket)
		return err
	}

	return retryHandler(b.bosClient, dlFunc, req, nil)
}

type putStorageClassReq struct {
	bucket       string
	storageClass string
}

func (p *putStorageClassReq) getBucketName() string {
	return p.bucket
}

// Wrapper of PutBucketStorageclass
func (b *bosClientWrapper) PutBucketStorageclass(bucket, storageClass string) error {

	req := &putStorageClassReq{
		bucket:       bucket,
		storageClass: storageClass,
	}

	plFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		plReq, ok := req.(*putStorageClassReq)
		if !ok {
			return fmt.Errorf("Error PutBucketStorageclass request type!")
		}
		err := bosClient.PutBucketStorageclass(plReq.bucket, plReq.storageClass)
		return err
	}

	return retryHandler(b.bosClient, plFunc, req, nil)
}

type getStorageClassReq struct {
	bucket string
}

func (g *getStorageClassReq) getBucketName() string {
	return g.bucket
}

type getStorageClassResp struct {
	ret string
}

// Wrapper of GetBucketStorageclass
func (b *bosClientWrapper) GetBucketStorageclass(bucket string) (string, error) {

	req := &getStorageClassReq{
		bucket: bucket,
	}
	resp := &getStorageClassResp{}

	gsFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		gsReq, ok := req.(*getStorageClassReq)
		if !ok {
			return fmt.Errorf("Error GetBucketStorageclass request type!")
		}
		gsResp, ok := resp.(*getStorageClassResp)
		if !ok {
			return fmt.Errorf("Error GetBucketStorageclass response type!")
		}
		ret, err := bosClient.GetBucketStorageclass(gsReq.bucket)
		if err == nil {
			gsResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, gsFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return "", err
}

type uploadPartCopyReq struct {
	srcBucket  string
	srcObject  string
	dstBucket  string
	dstObject  string
	uploadId   string
	partNumber int
	args       *api.UploadPartCopyArgs
}

func (c *uploadPartCopyReq) getBucketName() string {
	return c.dstBucket
}

type uploadPartCopyResp struct {
	ret *api.CopyObjectResult
}

// Wrapper of GetBucketStorageclass
func (b *bosClientWrapper) UploadPartCopy(bucket, object, srcBucket, srcObject, uploadId string,
	partNumber int, args *api.UploadPartCopyArgs) (*api.CopyObjectResult, error) {

	req := &uploadPartCopyReq{
		srcBucket:  srcBucket,
		srcObject:  srcObject,
		dstBucket:  bucket,
		dstObject:  object,
		uploadId:   uploadId,
		partNumber: partNumber,
		args:       args,
	}
	resp := &uploadPartCopyResp{}

	cpFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		cpReq, ok := req.(*uploadPartCopyReq)
		if !ok {
			return fmt.Errorf("Error uploadPartCopyReq request type!")
		}
		cpResp, ok := resp.(*uploadPartCopyResp)
		if !ok {
			return fmt.Errorf("Error uploadPartCopyResp response type!")
		}
		ret, err := bosClient.UploadPartCopy(cpReq.dstBucket, cpReq.dstObject, cpReq.srcBucket,
			cpReq.srcObject, cpReq.uploadId, cpReq.partNumber, cpReq.args)
		if err == nil {
			cpResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cpFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

// Wrapper of BasicUploadPart
func (b *bosClientWrapper) BasicUploadPart(bucket, object, uploadId string, partNumber int,
	content *bce.Body) (string, error) {

	var (
		ret         string
		err         error
		retryBuffer bytes.Buffer
	)

	teeReader := io.TeeReader(content.Stream(), &retryBuffer)
	content.SetStream(ioutil.NopCloser(teeReader))

	// try get endpoint from cache and execute function 'call'
	if ok := modifiyBosClientEndpointByBucketNameCache(b.bosClient, bucket); ok {
		ret, err = b.bosClient.BasicUploadPart(bucket, object, uploadId, partNumber, content)
		if err == nil || !shouldRetry(err) {
			return ret, err
		} else {
			ioutil.ReadAll(teeReader)
			content.SetStream(ioutil.NopCloser(&retryBuffer))
		}
	}

	// retry execute function 'call', but get endpiont from BOS
	log.Debugf("First process failed or can't get endpoint of bucket from cache, error: %s", err)
	epErr := modifiyBosClientEndpointByBucketNameBos(b.bosClient, bucket)
	if epErr == nil {
		return b.bosClient.BasicUploadPart(bucket, object, uploadId, partNumber, content)
	}
	if err != nil {
		return "", err
	}
	return "", epErr
}

type uploadPartWithBytesReq struct {
	bucket     string
	object     string
	uploadId   string
	partNumber int
	content    []byte
	args       *api.UploadPartArgs
}

func (c *uploadPartWithBytesReq) getBucketName() string {
	return c.bucket
}

type uploadPartWithBytesResp struct {
	ret string
}

// Wrapper of GetBucketStorageclass
func (b *bosClientWrapper) UploadPartFromBytes(bucket, object, uploadId string, partNumber int,
	content []byte, args *api.UploadPartArgs) (string, error) {

	req := &uploadPartWithBytesReq{
		bucket:     bucket,
		object:     object,
		uploadId:   uploadId,
		partNumber: partNumber,
		content:    content,
		args:       args,
	}
	resp := &uploadPartWithBytesResp{}

	cpFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		cpReq, ok := req.(*uploadPartWithBytesReq)
		if !ok {
			return fmt.Errorf("Error uploadPartWithBytesReq request type!")
		}
		cpResp, ok := resp.(*uploadPartWithBytesResp)
		if !ok {
			return fmt.Errorf("Error uploadPartWithBytesResp response type!")
		}
		ret, err := bosClient.UploadPartFromBytes(cpReq.bucket, cpReq.object, cpReq.uploadId,
			cpReq.partNumber, cpReq.content, cpReq.args)
		if err == nil {
			cpResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cpFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return "", err
}

type initiateMultipartUploadReq struct {
	bucket      string
	object      string
	contentType string
	args        *api.InitiateMultipartUploadArgs
}

func (c *initiateMultipartUploadReq) getBucketName() string {
	return c.bucket
}

type initiateMultipartUploadResp struct {
	ret *api.InitiateMultipartUploadResult
}

// Wrapper of GetBucketStorageclass
func (b *bosClientWrapper) InitiateMultipartUpload(bucket, object, contentType string,
	args *api.InitiateMultipartUploadArgs) (*api.InitiateMultipartUploadResult, error) {

	req := &initiateMultipartUploadReq{
		bucket:      bucket,
		object:      object,
		contentType: contentType,
		args:        args,
	}
	resp := &initiateMultipartUploadResp{}

	cpFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		inReq, ok := req.(*initiateMultipartUploadReq)
		if !ok {
			return fmt.Errorf("Error initiateMultipartUploadReq request type!")
		}
		inResp, ok := resp.(*initiateMultipartUploadResp)
		if !ok {
			return fmt.Errorf("Error initiateMultipartUploadResp response type!")
		}
		ret, err := bosClient.InitiateMultipartUpload(inReq.bucket, inReq.object, inReq.contentType,
			inReq.args)
		if err == nil {
			inResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cpFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type abortMultipartUploadReq struct {
	bucket   string
	object   string
	uploadId string
}

func (c *abortMultipartUploadReq) getBucketName() string {
	return c.bucket
}

func (b *bosClientWrapper) AbortMultipartUpload(bucket, object, uploadId string) error {
	req := &abortMultipartUploadReq{
		bucket:   bucket,
		object:   object,
		uploadId: uploadId,
	}

	amFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		amReq, ok := req.(*abortMultipartUploadReq)
		if !ok {
			return fmt.Errorf("Error abortMultipartUploadReq request type!")
		}
		err := bosClient.AbortMultipartUpload(amReq.bucket, amReq.object, amReq.uploadId)
		return err
	}
	return retryHandler(b.bosClient, amFunc, req, nil)
}

type completeMultipartUploadFromStructReq struct {
	bucket   string
	object   string
	uploadId string
	parts    *api.CompleteMultipartUploadArgs
}

func (c *completeMultipartUploadFromStructReq) getBucketName() string {
	return c.bucket
}

type completeMultipartUploadFromStructResp struct {
	ret *api.CompleteMultipartUploadResult
}

func (b *bosClientWrapper) CompleteMultipartUploadFromStruct(bucket, object, uploadId string,
	parts *api.CompleteMultipartUploadArgs) (*api.CompleteMultipartUploadResult, error) {

	req := &completeMultipartUploadFromStructReq{
		bucket:   bucket,
		object:   object,
		uploadId: uploadId,
		parts:    parts,
	}
	resp := &completeMultipartUploadFromStructResp{}

	cpFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		cpReq, ok := req.(*completeMultipartUploadFromStructReq)
		if !ok {
			return fmt.Errorf("Error completeMultipartUploadFromStructReq request type!")
		}
		cpResp, ok := resp.(*completeMultipartUploadFromStructResp)
		if !ok {
			return fmt.Errorf("Error completeMultipartUploadFromStructResp response type!")
		}
		ret, err := bosClient.CompleteMultipartUploadFromStruct(cpReq.bucket, cpReq.object,
			cpReq.uploadId, cpReq.parts)
		if err == nil {
			cpResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cpFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}

type getObjectReq struct {
	bucket          string
	object          string
	responseHeaders map[string]string
	ranges          []int64
}

func (g *getObjectReq) getBucketName() string {
	return g.bucket
}

type getObjectResp struct {
	ret *api.GetObjectResult
}

func (b *bosClientWrapper) GetObject(bucket, object string, responseHeaders map[string]string,
	ranges ...int64) (*api.GetObjectResult, error) {
	req := &getObjectReq{
		bucket:          bucket,
		object:          object,
		responseHeaders: responseHeaders,
		ranges:          ranges,
	}
	resp := &getObjectResp{}

	cpFunc := func(bosClient *bos.Client, req boscliReq, resp interface{}) error {
		goReq, ok := req.(*getObjectReq)
		if !ok {
			return fmt.Errorf("Error getObjectReq request type!")
		}
		goResp, ok := resp.(*getObjectResp)
		if !ok {
			return fmt.Errorf("Error getObjectResp response type!")
		}
		ret, err := bosClient.GetObject(goReq.bucket, goReq.object, goReq.responseHeaders,
			goReq.ranges...)
		if err == nil {
			goResp.ret = ret
		}
		return err
	}

	err := retryHandler(b.bosClient, cpFunc, req, resp)
	if err == nil {
		return resp.ret, nil
	}
	return nil, err
}
