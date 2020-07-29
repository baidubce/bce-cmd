## 0.3.0
  * 修复bug: bos sync fail时阻塞
  
## 0.2.9
  * 支持sts
  * 支持get-object-meta
  
## 0.2.8
  * 修复bug: 大文件cp时下载失败会假死
  
## 0.2.7
  * 修复bug: linux和mac下 sync的时候 a//b 正常匹配a/b
  
## 0.2.6
  * 修改 bos cp 命令usage说明文档
  * 修改逻辑: 获取 bucket to endpoint cache 失败时不需要退出。

## 0.2.5 
  * 修复BUG: 支持删除bos上的空目录
  * 修改bcecmd退出时的返回值：当批量上传或下载时，即使单个文件出错也以1退出。
  * 修复BUG: 当用户指定了 conf-path 时不在用户home目录下初始化配置文件

## 0.2.4
  * 优化断点续传做快照的速度
  * 修复BUG: head bucket 返回403时应该判定 bucket 存在
  * 修复BUG: 修复断点下载中的 bug
  * 为 sync 添加同步类型 time-size, time-size-crc32 和 only-crc32

## 0.2.3
  * 支持三步copy
  * 支持断点下载大文件
  * 为大文件 copy/uplaod/download 添加断点续传和进度条。
  * 支持前缀 bos://

## 0.2.2 

  * 修护BUG：为修改config添加读写锁。
  * 功能添加：支持用户手动为分支上传设置part size。
  * 更新go sdk： 修护上传400错误和body长度不同的问题。

## 0.2.1
	
*BOS CMD*

	* 修复BUG: 并行的时候bosClient的endpoint可能会混乱（添加读写锁）。
	* 更新gosdk到0.9.2。

## 0.2.0

*BOS CMD*

	*支持BOSAPI接口，包含Bucket ACL、生命周期、日志和默认存储类型等管理接口。
	*为sync 添加 exclude 和 include 过滤。
	* 修复Bug: 当关闭自动获取endpoint后，sync 和 copy时还是会自动获取endpoint。
	* 修复Bug：修复kingpin的若干Bug。
	* 修改kingpin 中help的输出模板， 1.使用户使用--help时显示的帮助文档更简洁；2.支持help换行。
